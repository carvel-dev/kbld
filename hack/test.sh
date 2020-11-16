#!/bin/bash

set -e -x -u

git --version || (echo "Missing git binary (used by tests)" && exit 1)

./hack/lint.sh

go clean -testcache

go test ./pkg/... -test.v $@

echo UNIT SUCCESS
