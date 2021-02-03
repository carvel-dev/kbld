#!/bin/bash

set -e -x -u

if [ "$(pack version)" != "v0.8.1 (git sha: e776ebf0096363bbac60771a456af941827316be)" ]; then
  echo "Please install 'pack' from https://github.com/buildpacks/pack/releases/tag/v0.8.1"
  exit 1
fi

if [ "$(kubectl-buildkit version)" != "refs/tags/v0.1.0" ]; then
  echo "Please install 'kubectl-buildkit' from https://github.com/vmware-tanzu/buildkit-cli-for-kubectl/releases/tag/v0.1.0"
  exit 1
fi

if [ "$(ko version)" != "0.8.0" ]; then
  echo "Please install 'ko' from https://github.com/google/ko/releases/tag/v0.8.0"
  exit 1
fi

go clean -testcache

export KBLD_BINARY_PATH="${KBLD_BINARY_PATH:-$PWD/kbld}"

go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS
