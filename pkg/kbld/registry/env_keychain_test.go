// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"testing"

	regname "github.com/google/go-containerregistry/pkg/name"
)

// Registry hostnames are case-insensitive DNS names (RFC 1035/1123). An
// operator-configured KBLD_REGISTRY_HOSTNAME_* value must resolve
// credentials for a target image reference even when the two differ only
// by case.
func TestEnvKeychainResolveIsCaseInsensitive(t *testing.T) {
	t.Setenv("KBLD_TEST_HOSTNAME", "MyRegistry.Example.com")
	t.Setenv("KBLD_TEST_USERNAME", "user")
	t.Setenv("KBLD_TEST_PASSWORD", "pass")

	k := NewEnvKeychain("KBLD_TEST")

	target, err := regname.NewRegistry("myregistry.example.com",
		regname.StrictValidation)
	if err != nil {
		t.Fatalf("building target registry: %s", err)
	}

	auth, err := k.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve returned error: %s", err)
	}

	cfg, err := auth.Authorization()
	if err != nil {
		t.Fatalf("Authorization returned error: %s", err)
	}
	if cfg.Username != "user" || cfg.Password != "pass" {
		t.Fatalf("case-differing hostname should resolve: got %+v",
			cfg)
	}
}

func TestEnvKeychainResolveNoMatchIsAnonymous(t *testing.T) {
	k := NewEnvKeychain("KBLD_TEST_UNSET")

	target, err := regname.NewRegistry("myregistry.example.com",
		regname.StrictValidation)
	if err != nil {
		t.Fatalf("building target registry: %s", err)
	}

	auth, err := k.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve returned error: %s", err)
	}
	cfg, err := auth.Authorization()
	if err != nil {
		t.Fatalf("Authorization returned error: %s", err)
	}
	if cfg.Username != "" || cfg.Password != "" {
		t.Fatalf("no env vars should give anonymous creds: got %+v",
			cfg)
	}
}
