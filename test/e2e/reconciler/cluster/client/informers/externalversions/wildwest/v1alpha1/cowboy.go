/*
Copyright 2021 The KCP Authors.

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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"

	wildwestv1alpha1 "github.com/kcp-dev/kcp/test/e2e/reconciler/cluster/apis/wildwest/v1alpha1"
	versioned "github.com/kcp-dev/kcp/test/e2e/reconciler/cluster/client/clientset/versioned"
	internalinterfaces "github.com/kcp-dev/kcp/test/e2e/reconciler/cluster/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kcp-dev/kcp/test/e2e/reconciler/cluster/client/listers/wildwest/v1alpha1"
)

// CowboyInformer provides access to a shared informer and lister for
// Cowboys.
type CowboyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CowboyLister
}

type cowboyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCowboyInformer constructs a new informer for Cowboy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCowboyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCowboyInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCowboyInformer constructs a new informer for Cowboy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCowboyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.WildwestV1alpha1().Cowboys(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.WildwestV1alpha1().Cowboys(namespace).Watch(context.TODO(), options)
			},
		},
		&wildwestv1alpha1.Cowboy{},
		resyncPeriod,
		indexers,
	)
}

func (f *cowboyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCowboyInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cowboyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&wildwestv1alpha1.Cowboy{}, f.defaultInformer)
}

func (f *cowboyInformer) Lister() v1alpha1.CowboyLister {
	return v1alpha1.NewCowboyLister(f.Informer().GetIndexer())
}
