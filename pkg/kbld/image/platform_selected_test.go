// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlimg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/image"
)

// TestMatchesPlatform runs test cases on the matchesPlatform function which verifies
// whether the given platform can run on the required platform by checking the
// compatibility of architecture, OS, OS version, OS features, variant and features.
// Adapted from https://github.com/google/go-containerregistry/blob/570ba6c88a5041afebd4599981d849af96f5dba9/pkg/v1/remote/index_test.go#L251
func TestMatchesPlatformSelection(t *testing.T) {
	tests := []struct {
		// want is the expected return value from matchesPlatform
		// when the given platform is 'given' and the required platform is 'required'.
		given    v1.Platform
		required ctlconf.PlatformSelection
		want     bool
	}{{ // The given & required platforms are identical. matchesPlatform expected to return true.
		given: v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
			OSVersion:    "10.0.10586",
			OSFeatures:   []string{"win32k"},
			Variant:      "armv6l",
			Features:     []string{"sse4"},
		},
		required: ctlconf.PlatformSelection{
			Architecture: "amd64",
			OS:           "linux",
			OSVersion:    "10.0.10586",
			OSFeatures:   []string{"win32k"},
			Variant:      "armv6l",
			Features:     []string{"sse4"},
		},
		want: true,
	},
		{ // OS and Architecture must exactly match. matchesPlatform expected to return false.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win32k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // OS version must exactly match
			given: v1.Platform{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10587",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // OS Features must exactly match. matchesPlatform expected to return false.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win32k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // Variant must exactly match. matchesPlatform expected to return false.
			given: v1.Platform{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv7l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // OS must exactly match, and is case sensative. matchesPlatform expected to return false.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "LinuX",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // OSVersion and Variant are specified in given but not in required.
			// matchesPlatform expected to return true.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{"win64k"},
				Variant:      "",
				Features:     []string{"sse4"},
			},
			want: true,
		},
		{ // Ensure the optional field OSVersion & Variant match exactly if specified as required.
			given: v1.Platform{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{},
				Variant:      "",
				Features:     []string{},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "amd64",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win32k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: false,
		},
		{ // Checking subset validity when required less features than given features.
			// matchesPlatform expected to return true.
			given: v1.Platform{
				Architecture: "",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win32k"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "",
				OS:           "linux",
				OSVersion:    "",
				OSFeatures:   []string{},
				Variant:      "",
				Features:     []string{},
			},
			want: true,
		},
		{ // Checking subset validity when required features are subset of given features.
			// matchesPlatform expected to return true.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k", "f1", "f2"},
				Variant:      "",
				Features:     []string{"sse4", "f1"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "",
				Features:     []string{"sse4"},
			},
			want: true,
		},
		{ // Checking subset validity when some required features is not subset of given features.
			// matchesPlatform expected to return false.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k", "f1", "f2"},
				Variant:      "",
				Features:     []string{"sse4", "f1"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k"},
				Variant:      "",
				Features:     []string{"sse4", "f2"},
			},
			want: false,
		},
		{ // Checking subset validity when OS features not required,
			// and required features is indeed a subset of given features.
			// matchesPlatform expected to return true.
			given: v1.Platform{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{"win64k", "f1", "f2"},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			required: ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "10.0.10586",
				OSFeatures:   []string{},
				Variant:      "armv6l",
				Features:     []string{"sse4"},
			},
			want: true,
		},
	}

	for _, test := range tests {
		got := ctlimg.MatchesPlatformSelection(test.given, test.required)
		if got != test.want {
			t.Errorf("matchesPlatform(%v, %v); got %v, want %v", test.given, test.required, got, test.want)
		}
	}
}
