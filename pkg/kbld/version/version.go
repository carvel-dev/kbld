// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"runtime/debug"
	"strings"
)

var (
	// Version can be set via:
	// -ldflags="-X 'carvel.dev/kbld/pkg/kbld/version.Version=$TAG'"
	defaultVersion = "develop"
	Version        = ""
	moduleName     = "carvel.dev/kbld"
)

func init() {
	Version = version()
}

func version() string {
	if Version != "" {
		// Version was set via ldflags, just return it.
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return defaultVersion
	}

	// Anything else.
	for _, dep := range info.Deps {
		if dep.Path == moduleName {
			// The minimumRequiredVersion field in kbld config is populated
			// from this version. Since the validation logic doesn't allow
			// minimumRequiredVersion to have a 'v' prefix, we remove it here.
			return strings.TrimPrefix(dep.Version, "v")
		}
	}

	return defaultVersion
}
