#!/bin/bash
set -e -x -u

go clean -testcache
./hack/build.sh
export KBLD_BINARY_PATH="${KBLD_BINARY_PATH:-$PWD/kbld}"

function build_test_deps() {
cat <<EOF | $KBLD_BINARY_PATH -f - | grep 'image:' | awk '{print $2}'
image: test-dependencies
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
sources:
- image: test-dependencies
  path: .
  docker:
    build:
      pull: true
      noCache: false
      file: hack/Dockerfile.dev
EOF
}

image_name=$(build_test_deps)

tempConfigFile=$(mktemp)
trap "rm -f $tempConfigFile" EXIT

minikube docker-env | while read env; do
  echo $env | grep -E 'export*' | awk '{print $2}' | sed 's/"//g'
done > $tempConfigFile

docker run \
--privileged \
--env-file $tempConfigFile \
-e KBLD_E2E_SKIP_WHEN_HTTP_REGISTRY=${KBLD_E2E_SKIP_WHEN_HTTP_REGISTRY:-true} \
-e KBLD_E2E_DOCKERHUB_HOSTNAME=`minikube ip`:30777 \
-v ~/.config:/root/.config \
-v ~/.minikube:"$HOME/.minikube" \
-v ~/.kube:/root/.kube \
-v /etc/docker/:/host-etc-docker \
--workdir /go/src/kbld \
-i -a STDOUT -a STDERR \
--network host --rm \
$image_name \
./hack/test-all-minikube-local-registry.sh $@