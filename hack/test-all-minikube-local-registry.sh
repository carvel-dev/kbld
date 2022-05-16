#!/bin/bash

set -e -x -u

export KBLD_E2E_DOCKERHUB_USERNAME=minikube-tests
export KBLD_E2E_DOCKERHUB_HOSTNAME=$(minikube ip):30777
# uncomment to disable stress tests
# export KBLD_E2E_SKIP_STRESS_TESTS=true

# Install secretgen to generate registry certs
kapp deploy -a sg -f https://github.com/vmware-tanzu/carvel-secretgen-controller/releases/download/v0.8.0/release.yml -y
# Install local docker2 registry
kapp deploy -a reg -f <(ytt -f test/e2e/assets/minikube-local-registry.yml -v registry_alt_name=$(minikube ip)) -y

if [ ! -z "$GITHUB_RUN_ID" ]; then
  # Install registry ca cert on the host machine
  kubectl get secret registry-ca-cert -ojsonpath='{.data.crt\.pem}' | base64 --decode > registry-ca-cert.crt
  sudo cp registry-ca-cert.crt /usr/local/share/ca-certificates/
  sudo update-ca-certificates

  # Docker needs its own CA cert configuration
  sudo mkdir -p /etc/docker/certs.d/${KBLD_E2E_DOCKERHUB_HOSTNAME}
  sudo cp registry-ca-cert.crt /etc/docker/certs.d/${KBLD_E2E_DOCKERHUB_HOSTNAME}/ca.crt

  sudo systemctl restart docker
  # Need to bootstrap to avoid race conditions to boot
  docker buildx create minikube --use --driver=kubernetes --bootstrap
fi

./hack/test-all.sh $@
