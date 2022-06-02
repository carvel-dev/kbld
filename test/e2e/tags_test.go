//go:build e2e

// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestAdditionalImageTags(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kbld := Kbld{t, env.KbldBinaryPath, logger}

	var published map[string]interface{}

	logger.Section("build and tag image", func() {
		input := env.WithRegistries(`
image: docker.io/*username*/kbld-e2e-tests-build
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/*username*/kbld-e2e-tests-build
  path: assets/simple-app
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: docker.io/*username*/kbld-e2e-tests-build
  tags:
  - staging
  - v1.0.0
`)

		out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
			StdinReader: strings.NewReader(input),
		})

		maskedOut := strings.Replace(out, regexp.MustCompile("sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED", -1)

		expectedOut := env.WithRegistries(`---
image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED
`)

		if maskedOut != expectedOut {
			t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", maskedOut, expectedOut)
		}

		err := yaml.Unmarshal([]byte(out), &published)
		if err != nil {
			t.Fatalf("Parsing output")
		}
	})

	logger.Section("check that pushed tags have published digest", func() {
		input := env.WithRegistries(`
kind: Object
spec:
- image: docker.io/*username*/kbld-e2e-tests-build:staging
- image: docker.io/*username*/kbld-e2e-tests-build:v1.0.0
`)

		out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
			StdinReader: strings.NewReader(input),
		})

		expectedOut := fmt.Sprintf(`---
kind: Object
spec:
- image: %s
- image: %s
`, published["image"], published["image"])

		if out != expectedOut {
			t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
		}
	})
}
