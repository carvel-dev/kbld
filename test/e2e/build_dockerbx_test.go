//go:build e2e

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"strings"
	"testing"
)

func TestDockerBuildxBuildSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := env.WithRegistries(`
kind: Object
spec:
- image: simple-app-two
- image: simple-app-three
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: simple-app-two
  path: assets/simple-app
  docker:
    buildx:
      file: dev/Dockerfile.dev
- image: simple-app-three
  path: assets/simple-app
  docker:
    buildx:
      target: build-env
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = regexp.MustCompile("sha256\\-[a-z0-9]{64}").ReplaceAllString(out, "SHA256-REPLACED")

	expectedOut := `---
kind: Object
spec:
- image: kbld:simple-app-two-SHA256-REPLACED
- image: kbld:simple-app-three-SHA256-REPLACED
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestDockerBuildxBuildAndPushSuccessful(t *testing.T) {
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
  docker:
    buildx: {}
- image: docker.io/*username*/kbld-e2e-tests-build2
  path: assets/simple-app
  docker:
    buildx:
      # try out multi platform build
      rawOptions: ["--platform=linux/amd64,linux/arm64,linux/arm/v7"]
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

	out = strings.Replace(out, regexp.MustCompile("sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED1", -1)
	out = strings.Replace(out, regexp.MustCompile("sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED2", -1)

	expectedOut := env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED1
- image: index.docker.io/*username*/kbld-e2e-tests-build2@SHA256-REPLACED2
`)

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
