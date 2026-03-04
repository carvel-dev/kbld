//go:build e2e

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestBazelBuildAndPushSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	assetPath := "assets/simple-app"
	secondAssetPath := "assets/simple-app-2"
	if err := exec.Command("cp", "-r", assetPath, secondAssetPath).Run(); err != nil {
		t.Fatalf("failed to copy %s to %s: %v", assetPath, secondAssetPath, err)
	}
	defer func() {
		if err := exec.Command("rm", "-rf", secondAssetPath).Run(); err != nil {
			t.Logf("failed to remove %s: %v", secondAssetPath, err)
		}
	}()

	input := env.WithRegistries(fmt.Sprintf(`
kind: Object
spec:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: docker.io/*username*/kbld-e2e-tests-build2
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/*username*/kbld-e2e-tests-build
  path: %s
  bazel:
    run:
      target: :simple-app
- image: docker.io/*username*/kbld-e2e-tests-build2
  path: %s
  bazel:
    run:
      target: :simple-app
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: docker.io/*username*/kbld-e2e-tests-build2
`, assetPath, secondAssetPath))

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = strings.Replace(out, regexp.MustCompile("sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED", -1)

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
