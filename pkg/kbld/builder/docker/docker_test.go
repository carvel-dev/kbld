// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"testing"
)

func TestLowerCaseRepository(t *testing.T) {
	testCases := lowerCaseRepositoryTestCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lowerCaseRepository(tc.input)
			if result != tc.expected {
				t.Errorf(
					"lowerCaseRepository(%q) = %q, expected %q",
					tc.input, result, tc.expected)
			}
		})
	}
}

func lowerCaseRepositoryTestCases() []struct {
	name     string
	input    string
	expected string
} {
	return []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase hostname, preserve path case",
			input:    "Harbor.Example.COM/MyApp/Image:v1.0",
			expected: "harbor.example.com/MyApp/Image:v1.0",
		},
		{
			name:     "mixed case registry and repository",
			input:    "ghcr.io/MyOrg/MyRepo:latest",
			expected: "ghcr.io/MyOrg/MyRepo:latest",
		},
		{
			name:     "uppercase registry only",
			input:    "REGISTRY.COM/lowercase/path:tag",
			expected: "registry.com/lowercase/path:tag",
		},
		{
			name:     "no tag specified",
			input:    "Registry.Example.COM/MyApp/Image",
			expected: "registry.example.com/MyApp/Image",
		},
		{
			name:     "digest reference",
			input:    "Harbor.Example.COM/MyApp/Image@sha256:abc123def456",
			expected: "harbor.example.com/myapp/image@sha256:abc123def456",
		},
		{
			name:     "localhost registry",
			input:    "LocalHost:5000/MyImage:v1",
			expected: "localhost:5000/MyImage:v1",
		},
		{
			name:     "case sensitive path preservation",
			input:    "registry.io/Path/To/MyImage:Tag",
			expected: "registry.io/Path/To/MyImage:Tag",
		},
		{
			name:     "bare image name with tag",
			input:    "MyImage:Tag",
			expected: "MyImage:Tag",
		},
		{
			name:     "bare image name without tag",
			input:    "MyImage",
			expected: "MyImage",
		},
		{
			name:     "implicit registry with namespace",
			input:    "MyOrg/MyImage:Tag",
			expected: "MyOrg/MyImage:Tag",
		},
		{
			name:     "malformed digest with implicit namespace",
			input:    "MyOrg/MyImage@sha256:deadbeef",
			expected: "myorg/myimage@sha256:deadbeef",
		},
		{
			name:     "uppercase localhost",
			input:    "LOCALHOST:5000/MyImage:Tag",
			expected: "localhost:5000/MyImage:Tag",
		},
		{
			name:     "explicit docker.io with uppercase repository",
			input:    "Docker.io/MyOrg/MyImage:Tag",
			expected: "docker.io/MyOrg/MyImage:Tag",
		},
		{
			name:     "DOCKER.IO with uppercase repository",
			input:    "DOCKER.IO/MyOrg/MyImage:Tag",
			expected: "docker.io/MyOrg/MyImage:Tag",
		},
		{
			name:     "implicit registry with index.docker.io in path",
			input:    "MyOrg/index.docker.io-mirror:Tag",
			expected: "MyOrg/index.docker.io-mirror:Tag",
		},
	}
}
