// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package image

import "strings"

// RegistryHostEqual compares explicit registry hosts case-insensitively without
// changing repository paths, tags, digests, or implicit Docker Hub namespaces.
func RegistryHostEqual(a, b string) bool {
	return normalizeRegistryHost(a) == normalizeRegistryHost(b)
}
func normalizeRegistryHost(ref string) string {
	host, rest, found := strings.Cut(ref, "/")
	explicit := strings.ContainsAny(host, ".:") ||
		strings.EqualFold(host, "localhost")
	if !found || !explicit {
		return ref
	}
	return strings.ToLower(host) + "/" + rest
}
