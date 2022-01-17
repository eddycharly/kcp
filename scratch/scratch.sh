#!/usr/bin/env bash

set -e

export KUBECONFIG=.kcp/admin.kubeconfig
export LOCALIP=$(ipconfig getifaddr en0)

deleteCluster() {
  kind delete cluster --name $1 --kubeconfig $1.kubeconfig
}

createCluster() {
  kind create cluster --name $1 --kubeconfig $1.kubeconfig --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "127.0.0.1"
  ipFamily: ipv4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
EOF
}

registerCluster() {
(cat <<-EOF
apiVersion: cluster.example.dev/v1alpha1
kind: Cluster
metadata:
  name: $1
spec:
  kubeconfig: |
EOF
sed -e 's/^/    /' $1.kubeconfig) | kubectl apply -f -
}

startKcp() {
  go run ../cmd/kcp start --listen=${LOCALIP}:6443
}

setupCRDs() {
  kubectl apply -f ../config/
  # kubectl apply -f ../contrib/crds/apps
}

rm *.kubeconfig
rm -rf .kcp

deleteCluster a &
deleteCluster b &
deleteCluster c &

wait

createCluster a &
createCluster b &
createCluster c &

wait

startKcp &
sleep 40s
setupCRDs
registerCluster a
registerCluster b
registerCluster c

wait

git fetch upstream
git checkout main
git rebase upstream/main
git push -f origin main
