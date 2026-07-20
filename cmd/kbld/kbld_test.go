// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const failureExitCode = 1

func TestJSONErrorsAreWrittenToStderr(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestKbldProcess$", "--",
		"--json", "-f", filepath.Join(t.TempDir(), "missing.yml"))
	cmd.Env = append(os.Environ(), "KBLD_TEST_PROCESS=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, failureExitCode, exitErr.ExitCode())
	assert.Empty(t, stdout.String())
	assert.True(t, strings.HasPrefix(
		stderr.String(), "kbld: Error: Unable to stat file:"))
}

func TestKbldProcess(t *testing.T) {
	if os.Getenv("KBLD_TEST_PROCESS") != "1" {
		return
	}

	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"kbld"}, os.Args[i+1:]...)
			main()
			return
		}
	}

	t.Fatal("missing command separator")
}
