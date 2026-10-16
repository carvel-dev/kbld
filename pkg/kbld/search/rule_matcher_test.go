// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package search_test

import (
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlsearch "carvel.dev/kbld/pkg/kbld/search"
)

const (
	lowerCaseDockerIoImg = "docker.io/img"
	nonStringTestValue   = 123
)

// Registry hostnames are case-insensitive (RFC 1035/1123).
// ValueMatcher.Image / ValueMatcher.ImageRepo come from the user's
// kbld.yaml config while the scanned value comes from the target
// manifest, so a hostname-case mismatch between the two sources
// must not prevent a match.
func TestRuleMatcherValueMatcherCaseInsensitiveHostname(t *testing.T) {
	type example struct {
		desc    string
		rule    ctlconf.SearchRule
		value   any
		matched bool
	}

	imageMatcher := func(image string) ctlconf.SearchRule {
		return ctlconf.SearchRule{
			ValueMatcher: &ctlconf.SearchRuleValueMatcher{Image: image},
		}
	}
	repoMatcher := func(repo string) ctlconf.SearchRule {
		return ctlconf.SearchRule{
			ValueMatcher: &ctlconf.SearchRuleValueMatcher{
				ImageRepo: repo,
			},
		}
	}

	exs := []example{
		{
			desc:    "image: mismatched hostname case still matches",
			rule:    imageMatcher("Docker.IO/img"),
			value:   lowerCaseDockerIoImg,
			matched: true,
		},
		{
			desc:    "image: same case matches (regression guard)",
			rule:    imageMatcher(lowerCaseDockerIoImg),
			value:   lowerCaseDockerIoImg,
			matched: true,
		},
		{
			desc:    "image: repo path case remains significant",
			rule:    imageMatcher("docker.io/IMG"),
			value:   lowerCaseDockerIoImg,
			matched: false,
		},
		{
			desc:    "image: non-string value falls back to exact cmp",
			rule:    imageMatcher(lowerCaseDockerIoImg),
			value:   nonStringTestValue,
			matched: false,
		},
		{
			desc:    "imageRepo: mismatched hostname case still matches",
			rule:    repoMatcher("DOCKER.IO/img"),
			value:   "docker.io/img:tag",
			matched: true,
		},
		{
			desc:    "imageRepo: repo path case remains significant",
			rule:    repoMatcher("docker.io/IMG"),
			value:   lowerCaseDockerIoImg,
			matched: false,
		},
	}

	for _, ex := range exs {
		matcher := ctlsearch.NewRulesMatcher([]ctlconf.SearchRule{ex.rule})

		matched, _ := matcher.Matches(nil, ex.value)
		if matched != ex.matched {
			t.Errorf("%s: expected matched=%v, got %v",
				ex.desc, ex.matched, matched)
		}
	}
}
