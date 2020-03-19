package image_test

import (
	"testing"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
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
			ImageRef: ctlconf.ImageRef{ImageRepo: "index.docker.io/img"},
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
			ImageRef: ctlconf.ImageRef{ImageRepo: "index.docker.io/img"},
			URL:      "docker.io/img:tag@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
			Matched:  true,
		},
	}

	for _, ex := range exs {
		if ctlimg.NewMatcher(ex.URL).Matches(ex.ImageRef) != ex.Matched {
			t.Fatalf("Expected %#v to succeed", ex)
		}
	}
}
