// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"strings"
	"testing"
)

func TestInvalidRegistryHostnameDoesNotExposeCredentials(t *testing.T) {
	for _, suffix := range []string{"", "_0"} {
		t.Run("hostname"+suffix, func(t *testing.T) {
			raw := "https://user:password-canary@example.com:bad" +
				"?token=query-canary"
			t.Setenv("KBLD_REGISTRY_HOSTNAME"+suffix, raw)
			_, err := NewEnvKeychain("KBLD_REGISTRY").collect()
			if err == nil || strings.Contains(err.Error(), "canary") {
				t.Fatal("registry validation exposed input")
			}
		})
	}
}
