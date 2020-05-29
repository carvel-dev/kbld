#!/bin/bash

set -e -x -u

export KBLD_E2E_DOCKERHUB_USERNAME=minikube-tests
export KBLD_E2E_DOCKERHUB_HOSTNAME=$(minikube ip):30777
# export KBLD_E2E_SKIP_CF_IMAGES_DOWNLOAD=true

kapp deploy -a reg -f test/e2e/assets/minikube-local-registry.yml -y

./hack/test-all.sh $@
