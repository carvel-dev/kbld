#!/bin/bash

set -e -x -u

GO111MODULE=off go get -u github.com/kisielk/errcheck
# add -verbose to see the resulting list of added excludes.
errcheck -exclude "$PWD/hack/errcheck_excludes.txt" ./pkg/kbld/...
