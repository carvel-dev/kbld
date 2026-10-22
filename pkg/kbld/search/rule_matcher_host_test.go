// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package search

import (
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	"github.com/stretchr/testify/require"
)

func TestRulesMatchRegistryHostCase(t *testing.T) {
	refs := []ctlconf.SearchRuleValueMatcher{
		{Image: "REGISTRY.example.com/team/image:Tag"},
		{ImageRepo: "REGISTRY.example.com/team/image"},
	}
	for _, ref := range refs {
		matcher := RuleMatcher{rule: ctlconf.SearchRule{ValueMatcher: &ref}}
		matched, _ := matcher.Matches(nil,
			"registry.example.com/team/image:Tag")
		require.True(t, matched)
		matched, _ = matcher.Matches(nil,
			"registry.example.com/Team/image:Tag")
		require.False(t, matched)
	}
}

func TestFullImageRulePreservesTagCase(t *testing.T) {
	ref := ctlconf.SearchRuleValueMatcher{
		Image: "REGISTRY.example.com/team/image:Tag",
	}
	matcher := RuleMatcher{rule: ctlconf.SearchRule{ValueMatcher: &ref}}
	matched, _ := matcher.Matches(nil,
		"registry.example.com/team/image:tag")
	require.False(t, matched)
}
