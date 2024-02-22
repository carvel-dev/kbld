// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package logger_test

import (
	"bytes"
	"testing"

	ctllog "carvel.dev/kbld/pkg/kbld/logger"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := ctllog.NewLogger(&buf)
	prefLogger := logger.NewPrefixedWriter("prefix: ")

	prefLogger.Write([]byte("content1"))
	prefLogger.Write([]byte("content2\n"))
	prefLogger.Write([]byte("content3\ncontent4"))
	prefLogger.Write([]byte("content5\ncontent6\n"))
	prefLogger.Write([]byte("\ncontent7\ncontent8\n"))
	prefLogger.Write([]byte("\n\n"))

	out := buf.String()
	expectedOut := `prefix: content1
prefix: content2
prefix: content3
prefix: content4
prefix: content5
prefix: content6
prefix: 
prefix: content7
prefix: content8
prefix: 
prefix: 
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
