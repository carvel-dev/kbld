// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlimg "carvel.dev/kbld/pkg/kbld/image"
)

func TestMatcherMatches(t *testing.T) {
	type matcherExample struct {
		ctlconf.ImageRef
		URL     string
		Matched bool
	}

	exs := []matcherExample{
		// Match by image
		{
			ImageRef: ctlconf.ImageRef{Image: "img"},
			URL:      "img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "img/img"},
			URL:      "img/img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "docker.io/img"},
			URL:      "docker.io/img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "docker.io/img:tag"},
			URL:      "docker.io/img:tag",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "docker.io/img@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d"},
			URL:      "docker.io/img@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "docker.io/img:tag@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d"},
			URL:      "docker.io/img:tag@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		},

		// Match by image repo
		{
			ImageRef: ctlconf.ImageRef{ImageRepo: "img"},
			URL:      "img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "docker.io/img"},
			URL:      "docker.io/img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "docker.io/img"},
			URL:      "docker.io/img:tag",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "img"},
			URL:      "img:tag",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "docker.io/img"},
			URL:      "docker.io/img@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "docker.io/img"},
			URL:      "docker.io/img:tag@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "localhost:3000/org/img"},
			URL:      "localhost:3000/org/img:tag@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "localhost:3000/org/img"},
			URL:      "localhost:3000/org/img:tag",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "localhost:3000/org/img"},
			URL:      "localhost:3000/org/img",
			Matched:  true,
		},

		// Normalization is not supported
		{
			ImageRef: ctlconf.ImageRef{ImageRepo: "index.docker.io/img"},
			URL:      "docker.io/img",
			Matched:  false,
		},

		// Registry hostname is case-insensitive (RFC 1035/1123): a
		// mismatched hostname case between the manifest-scanned ImageRef and
		// the user-authored config URL must still match.
		{
			ImageRef: ctlconf.ImageRef{Image: "Docker.IO/img"},
			URL:      "docker.io/img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{Image: "docker.io/img"},
			URL:      "DOCKER.IO/img",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "DOCKER.IO/img"},
			URL:      "docker.io/img:tag",
			Matched:  true,
		}, {
			ImageRef: ctlconf.ImageRef{ImageRepo: "Localhost:3000/org/img"},
			URL:      "localhost:3000/org/img:tag",
			Matched:  true,
		},

		// Repository path case remains significant (documented design,
		// unlike the registry hostname segment above).
		{
			ImageRef: ctlconf.ImageRef{ImageRepo: "docker.io/IMG"},
			URL:      "docker.io/img",
			Matched:  false,
		},
	}

	for _, ex := range exs {
		if ctlimg.NewMatcher(ex.URL).Matches(ex.ImageRef) != ex.Matched {
			t.Fatalf("Expected %#v to succeed", ex)
		}
	}
}
