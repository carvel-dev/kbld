#!/bin/bash

set -e -x -u

LATEST_GIT_TAG=$(git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+')
VERSION="${1:-$LATEST_GIT_TAG}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X github.com/k14s/kbld/pkg/kbld/version.Version=$VERSION -buildid="


go fmt ./cmd/... ./pkg/... ./test/...
go mod vendor
go mod tidy

# export GOOS=linux GOARCH=amd64
go build -ldflags="$LDFLAGS" -trimpath -o kbld ./cmd/kbld/...
./kbld version

echo "Success"
