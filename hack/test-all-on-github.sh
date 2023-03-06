#!/bin/bash
set -e -x -u

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]:-$0}"; )" &> /dev/null && pwd 2> /dev/null; )";

go clean -testcache
./hack/build.sh

mkdir -p ~/.docker/cli-plugins/
curl -sLo ~/.docker/cli-plugins/docker-buildx https://github.com/docker/buildx/releases/download/v0.8.2/buildx-v0.8.2.linux-amd64
chmod +x ~/.docker/cli-plugins/docker-buildx
curl -sL https://github.com/buildpacks/pack/releases/download/v0.8.1/pack-v0.8.1-linux.tgz | tar -C /usr/local/bin -xz
curl -sL https://github.com/vmware-tanzu/buildkit-cli-for-kubectl/releases/download/v0.1.0/linux-refs.tags.v0.1.0.tgz | tar -C /usr/local/bin -xz
curl -sL https://github.com/google/ko/releases/download/v0.8.0/ko_0.8.0_Linux_x86_64.tar.gz | tar -C /usr/local/bin -xz
mkdir -p /usr/local/bin
curl -sLo /usr/local/bin/bazel https://github.com/bazelbuild/bazel/releases/download/4.2.0/bazel-4.2.0-linux-x86_64
chmod +x /usr/local/bin/bazel
curl -sLo /usr/local/bin/kapp https://github.com/carvel-dev/kapp/releases/download/v0.48.0/kapp-linux-amd64
chmod +x /usr/local/bin/kapp
curl -sLo /usr/local/bin/ytt https://github.com/carvel-dev/ytt/releases/download/v0.41.1/ytt-linux-amd64
chmod +x /usr/local/bin/ytt
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor > /usr/share/keyrings/cloud.google.gpg
apt-get update -y
apt-get install google-cloud-cli -y
git config --global user.email "email@example.com"
git config --global user.name "Some Person"
curl -sLo /usr/local/bin/kubectl "https://dl.k8s.io/release/$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x /usr/local/bin/kubectl


tempConfigFile=$(mktemp)
trap "rm -f $tempConfigFile" EXIT

minikube docker-env | while read env; do
  echo $env | grep -E 'export*' | awk '{print $2}' | sed 's/"//g'
done > $tempConfigFile

export KBLD_E2E_SKIP_WHEN_HTTP_REGISTRY=${KBLD_E2E_SKIP_WHEN_HTTP_REGISTRY:-true}
export KBLD_E2E_DOCKERHUB_HOSTNAME=`minikube ip`:30777

./hack/test-all-minikube-local-registry.sh $@
