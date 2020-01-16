#!/bin/bash

set -e -x -u

BUILD_VALUES= ./hack/build.sh

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

GOOS=darwin GOARCH=amd64 go build $repro_flags -o kbld-darwin-amd64 ./cmd/kbld/...
GOOS=linux GOARCH=amd64 go build $repro_flags -o kbld-linux-amd64 ./cmd/kbld/...
GOOS=windows GOARCH=amd64 go build $repro_flags -o kbld-windows-amd64.exe ./cmd/kbld/...

shasum -a 256 ./kbld-*-amd64*
