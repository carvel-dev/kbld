//go:build e2e

// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"strings"
	"testing"
)

type Env struct {
	Namespace            string
	DockerHubUsername    string
	DockerHubHostname    string
	SkipStressTests      bool
	SkipWhenHTTPRegistry bool
	KbldBinaryPath       string
}

func BuildEnv(t *testing.T) Env {
	kbldPath := os.Getenv("KBLD_BINARY_PATH")
	if kbldPath == "" {
		kbldPath = "kbld"
	}

	env := Env{
		DockerHubUsername:    os.Getenv("KBLD_E2E_DOCKERHUB_USERNAME"),
		DockerHubHostname:    os.Getenv("KBLD_E2E_DOCKERHUB_HOSTNAME"),
		SkipStressTests:      os.Getenv("KBLD_E2E_SKIP_STRESS_TESTS") == "true",
		SkipWhenHTTPRegistry: os.Getenv("KBLD_E2E_SKIP_WHEN_HTTP_REGISTRY") == "true",
		KbldBinaryPath:       kbldPath,
	}
	env.Validate(t)
	return env
}

func (e Env) Validate(t *testing.T) {
	errStrs := []string{}

	if len(e.DockerHubUsername) == 0 {
		errStrs = append(errStrs, "Expected DockerHubUsername to be non-empty")
	}

	if len(errStrs) > 0 {
		t.Fatalf("%s", strings.Join(errStrs, "\n"))
	}
}

func (e Env) WithRegistries(input string) string {
	for _, hostname := range []string{"index.docker.io/", "docker.io/"} {
		newHostname := hostname
		if len(e.DockerHubHostname) > 0 {
			newHostname = e.DockerHubHostname + "/"
		}
		input = strings.Replace(input, hostname+"*username*/", newHostname+e.DockerHubUsername+"/", -1)
	}
	return input
}
