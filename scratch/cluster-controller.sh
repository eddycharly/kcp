#!/usr/bin/env bash

set -e

export KUBECONFIG=.kcp/admin.kubeconfig

loadImage() {
  kind load docker-image --name $1 $2
}

startClusterController() {
  go run ../cmd/cluster-controller --pull-mode --syncer-image=$1 --kubeconfig=.kcp/admin.kubeconfig --auto-publish-apis
  # go run ../cmd/cluster-controller --push-mode --syncer-image=$1 --kubeconfig=.kcp/admin.kubeconfig --auto-publish-apis
}

SYNCER_IMAGE=$(ko publish --local ../cmd/syncer)
loadImage a $SYNCER_IMAGE &
loadImage b $SYNCER_IMAGE &
loadImage c $SYNCER_IMAGE &

wait

startClusterController $SYNCER_IMAGE
