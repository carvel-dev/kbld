//go:build e2e

// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"
)

func TestPackBuildAndPushSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := env.WithRegistries(`
kind: Object
spec:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: docker.io/*username*/kbld-e2e-tests-build2
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/*username*/kbld-e2e-tests-build
  path: assets/simple-app
  pack: &pack
    build:
      builder: index.docker.io/cloudfoundry/cnb@sha256:83270cf59e8944be0c544e45fd45a5a1f4526d7936d488d2de8937730341618d
- image: docker.io/*username*/kbld-e2e-tests-build2
  path: assets/simple-app
  pack: &pack
    build:
      builder: index.docker.io/cloudfoundry/cnb@sha256:83270cf59e8944be0c544e45fd45a5a1f4526d7936d488d2de8937730341618d
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: docker.io/*username*/kbld-e2e-tests-build2
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	// Replace digest multiple times as pack does not always produce image with the same digest
	// Possibly related: https://github.com/buildpack/lifecycle/issues/181
	digestRegex := regexp.MustCompile("sha256:[a-z0-9]{64}")
	for {
		digestStr := digestRegex.FindString(out)
		if len(digestStr) == 0 {
			break
		}
		out = strings.Replace(out, digestStr, "SHA256-REPLACED", -1)
	}

	expectedOut := env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED
- image: index.docker.io/*username*/kbld-e2e-tests-build2@SHA256-REPLACED
`)

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
