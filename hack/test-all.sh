#!/bin/bash

set -e -x -u

./hack/build.sh $@

export KBLD_BINARY_PATH="$PWD/kbld"

./hack/test.sh
./hack/test-e2e.sh $@

echo ALL SUCCESS
