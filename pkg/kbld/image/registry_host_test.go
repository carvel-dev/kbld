// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	"github.com/stretchr/testify/require"
)

func TestRegistryHostMatching(t *testing.T) {
	const imageWithTag = "team/image:Tag"
	host := "registry.example.com/"
	upper := "REGISTRY.example.com/"
	for _, tc := range []struct {
		a, b  string
		equal bool
	}{
		{upper + imageWithTag, host + imageWithTag, true},
		{"LOCALHOST:5000/image:Tag", "localhost:5000/image:Tag", true},
		{upper + imageWithTag, host + "team/image:tag", false},
		{host + "Team/image:Tag", host + imageWithTag, false},
		{host + "image@sha256:AB", host + "image@sha256:ab", false},
		{"Namespace/image", "namespace/image", false},
	} {
		require.Equal(t, tc.equal, RegistryHostEqual(tc.a, tc.b))
	}
}

func TestConfigMatchesRegistryHostCase(t *testing.T) {
	refs := []ctlconf.ImageRef{
		{Image: "REGISTRY.example.com/team/image:Tag"},
		{ImageRepo: "REGISTRY.example.com/team/image"},
	}
	for _, ref := range refs {
		matcher := NewMatcher("registry.example.com/team/image:Tag")
		require.True(t, matcher.Matches(ref))
	}
}

func TestHostCaseDoesNotDuplicatePreresolvedOrigins(t *testing.T) {
	original := []ctlconf.Origin{{Preresolved: &ctlconf.OriginPreresolved{
		URL: "REGISTRY.example.com/team/image:Tag",
	}}}
	image := NewPreresolvedImage("registry.example.com/team/image:Tag",
		original)
	_, origins, err := image.URL()
	require.NoError(t, err)
	require.Equal(t, original, origins)
}
