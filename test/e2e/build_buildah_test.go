//go:build e2e

// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestBuildahBuildAndPush(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Buildah is only available for linux, so we are skipping this test")
	}

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
  buildah:
    pull: true
- image: docker.io/*username*/kbld-e2e-tests-build2
  path: assets/simple-app
  buildah:
    # try out multi platform build
    platforms: ["linux/amd64","linux/arm64"]

---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: docker.io/*username*/kbld-e2e-tests-build2
  tags:
    - test
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = strings.Replace(out, regexp.MustCompile(
		"sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED1", -1)
	out = strings.Replace(out, regexp.MustCompile(
		"sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED2", -1)

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
