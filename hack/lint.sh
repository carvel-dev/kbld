#!/bin/bash

set -e -x -u

# add -verbose to see the resulting list of added excludes.
go run github.com/kisielk/errcheck -exclude "$PWD/hack/errcheck_excludes.txt" ./pkg/kbld/...
