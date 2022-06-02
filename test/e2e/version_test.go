//go:build e2e

// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	out, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})

	if !strings.Contains(out, "kbld version") {
		t.Fatalf("Expected to find client version")
	}
}
