#!/bin/bash

set -e -x -u

export KBLD_E2E_DOCKERHUB_USERNAME=minikube-tests
export KBLD_E2E_DOCKERHUB_HOSTNAME=${KBLD_E2E_DOCKERHUB_HOSTNAME:-`minikube ip`}
# uncomment to disable stress tests
# export KBLD_E2E_SKIP_STRESS_TESTS=true

# Install secretgen to generate registry certs
kapp deploy -a sg -f https://github.com/vmware-tanzu/carvel-secretgen-controller/releases/download/v0.8.0/release.yml -y
# Install local docker2 registry
kapp deploy -a reg -f <(ytt -f test/e2e/assets/minikube-local-registry.yml -v registry_alt_name=$(echo "$KBLD_E2E_DOCKERHUB_HOSTNAME" | cut -d: -f1)) -y

# Install registry ca cert on the host machine
kubectl get secret registry-ca-cert -ojsonpath='{.data.crt\.pem}' | base64 --decode > registry-ca-cert.crt
cp registry-ca-cert.crt /usr/local/share/ca-certificates/
update-ca-certificates

# Docker needs its own CA cert configuration
mkdir -p /host-etc-docker/certs.d/${KBLD_E2E_DOCKERHUB_HOSTNAME}
cp registry-ca-cert.crt /host-etc-docker/certs.d/${KBLD_E2E_DOCKERHUB_HOSTNAME}/ca.crt

# Buildkit needs to talk to above registry however
# it does not seem to properly auto-copy CA certificates
# so disable certificate verification
cat <<EOF >buildkitd.toml
[registry."${KBLD_E2E_DOCKERHUB_HOSTNAME}"]
  insecure = true
EOF

# Need to bootstrap to avoid race conditions to boot
docker buildx create minikube --use --driver=kubernetes --bootstrap --config buildkitd.toml

if [ "$2" == "github-workflow" ]
then
  git init . --bare
  git config --global --add safe.directory /__w/go/src/kbld
fi

./hack/test-all.sh $@
