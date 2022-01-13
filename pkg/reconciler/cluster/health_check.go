/*
Copyright 2022 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	clusterv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/cluster/v1alpha1"
	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	clusterinformer "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/cluster/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/client/listers/cluster/v1alpha1"
	coordinationinformers "k8s.io/client-go/informers/coordination/v1"
)

type HealthCheckController struct {
	queue          workqueue.RateLimitingInterface
	kcpClient      kcpclient.Interface
	clusterIndexer cache.Indexer
	clusterLister  v1alpha1.ClusterLister
	leaseIndexer   cache.Indexer
	syncChecks     []cache.InformerSynced
}

func NewHealthCheckController(
	kcpClient kcpclient.Interface,
	clusterInformer clusterinformer.ClusterInformer,
	leaseInformer coordinationinformers.LeaseInformer,
) (*HealthCheckController, error) {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &HealthCheckController{
		queue:          queue,
		kcpClient:      kcpClient,
		clusterIndexer: clusterInformer.Informer().GetIndexer(),
		clusterLister:  clusterInformer.Lister(),
		leaseIndexer:   leaseInformer.Informer().GetIndexer(),
		syncChecks: []cache.InformerSynced{
			clusterInformer.Informer().HasSynced,
			leaseInformer.Informer().HasSynced,
		},
	}

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.enqueue(obj) },
		UpdateFunc: func(_, obj interface{}) { c.enqueue(obj) },
	})

	leaseInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.enqueue(obj) },
		UpdateFunc: func(_, obj interface{}) { c.enqueue(obj) },
	})

	return c, nil
}

func (c *HealthCheckController) enqueue(obj interface{}) {
	c.queue.Add(obj)
}

func (c *HealthCheckController) Start(ctx context.Context, numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting Cluster health check controller")
	defer klog.Info("Shutting down Cluster health check controller")

	if !cache.WaitForNamedCacheSync("clusterhealth", ctx.Done(), c.syncChecks...) {
		klog.Warning("Failed to wait for caches to sync")
		return
	}

	for i := 0; i < numThreads; i++ {
		go wait.Until(func() { c.startWorker(ctx) }, time.Second, ctx.Done())
	}

	<-ctx.Done()
}

func (c *HealthCheckController) startWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *HealthCheckController) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	key, _ := meta.Accessor(k)

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	if err := c.process(ctx, key); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller failed to sync %q, err: %w", controllerName, key, err))
		c.queue.AddRateLimited(key)
		return true
	}
	c.queue.Forget(key)
	return true
}

func (c *HealthCheckController) process(ctx context.Context, key metav1.Object) error {
	k, err := cache.MetaNamespaceKeyFunc(key)
	if err != nil {
		runtime.HandleError(err)
		return err
	}
	obj, err := c.clusterLister.Get(k)
	// obj, exists, err := c.clusterIndexer.GetByKey(key.GetName())
	if err != nil {
		return err
	}

	// if !exists {
	// 	klog.Errorf("Object with key %q was deleted", key)
	// 	return nil
	// }
	current := obj.DeepCopy()
	// current := obj.(*clusterv1alpha1.Cluster).DeepCopy()
	previous := current.DeepCopy()

	if err := c.reconcile(ctx, current); err != nil {
		return err
	}

	// If the object being reconciled changed as a result, update it.
	if !equality.Semantic.DeepEqual(previous.Status, current.Status) {
		_, uerr := c.kcpClient.ClusterV1alpha1().Clusters().UpdateStatus(ctx, current, metav1.UpdateOptions{})
		return uerr
	}

	return nil
}

func (c *HealthCheckController) reconcile(ctx context.Context, cluster *clusterv1alpha1.Cluster) error {
	klog.Infof("reconciling cluster health %q", cluster.Name)

	// logicalCluster := cluster.GetClusterName()
	return nil
}
