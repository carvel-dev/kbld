#!/bin/bash

set -e -x -u

if [ "$(pack version)" != "v0.8.1 (git sha: e776ebf0096363bbac60771a456af941827316be)" ]; then
  echo "Please install 'pack' from https://github.com/buildpacks/pack/releases/tag/v0.8.1"
  exit 1
fi

go clean -testcache

export KBLD_BINARY_PATH="${KBLD_BINARY_PATH:-$PWD/kbld}"

go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS
