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

func TestPackBuildAndPushSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	// Copy asset to avoid parallel build interference
	assetPath := "assets/simple-app"
	secondAssetPath := "assets/simple-app-2"
	copyCmd := exec.Command("cp", "-r", assetPath, secondAssetPath)
	if err := copyCmd.Run(); err != nil {
		t.Fatalf("failed to copy asset from %s to %s: %v", assetPath, secondAssetPath, err)
	}
	defer func() {
		if err := exec.Command("rm", "-rf", secondAssetPath).Run(); err != nil {
			t.Logf("failed to remove temporary asset path %s: %v", secondAssetPath, err)
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
  pack: &pack
    build:
      builder: index.docker.io/cloudfoundry/cnb@sha256:83270cf59e8944be0c544e45fd45a5a1f4526d7936d488d2de8937730341618d
- image: docker.io/*username*/kbld-e2e-tests-build2
  path: %s
  pack: &pack
    build:
      builder: index.docker.io/cloudfoundry/cnb@sha256:83270cf59e8944be0c544e45fd45a5a1f4526d7936d488d2de8937730341618d
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
