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

# Ensure the host kernel (Minikube VM) is configured to run cross-platform binaries.
# This is required for Buildah to execute steps (like RUN) for linux/arm64 on an amd64 host.
docker run --privileged --rm tonistiigi/binfmt --install all

minikube docker-env | while read env; do
  echo $env | grep -E 'export*' | awk '{print $2}' | sed 's/"//g'
done > $tempConfigFile

docker run \
--privileged \
--env-file $tempConfigFile \
-e DOCKER_API_VERSION=1.44 \
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