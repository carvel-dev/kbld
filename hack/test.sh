#!/bin/bash

set -e -x -u

git --version || (echo "Missing git binary (used by tests)" && exit 1)

go clean -testcache

go test ./... -test.v $@

echo UNIT SUCCESS
