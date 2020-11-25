#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

go fmt ./cmd/... ./pkg/... ./test/...
go mod vendor
go mod tidy

build_values_path="../../../${BUILD_VALUES:-./hack/build-values-default.yml}"

(
	# template all playground assets
	# into a single Go file
	cd pkg/kbld/website; 

	ytt version || { echo >&2 "ytt is required for building. Install from https://github.com/vmware-tanzu/carvel-ytt"; exit 1; }
	ytt -f . -f $build_values_path \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--dangerous-emptied-output-directory ../../../tmp/
)
mv tmp/generated.go.txt pkg/kbld/website/generated.go

# export GOOS=linux GOARCH=amd64
go build $repro_flags -o kbld ./cmd/kbld/...
./kbld version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build $repro_flags -o ./tmp/main ./cmd/kbld-lambda-website/...
(
	cd tmp
	chmod +x main
	rm -f kbld-lambda-website.zip
	zip kbld-lambda-website.zip main
)

echo "Success"
