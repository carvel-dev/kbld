// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package search

import (
	"reflect"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlimg "carvel.dev/kbld/pkg/kbld/image"
	ctlres "carvel.dev/kbld/pkg/kbld/resources"
)

type Matcher interface {
	Matches(keyPath ctlres.Path, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy)
}

type RuleMatcher struct {
	rule ctlconf.SearchRule
}

var _ Matcher = RuleMatcher{}

func (m RuleMatcher) Matches(keyPath ctlres.Path, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy) {
	var keyMatched, valueMatched bool

	if m.rule.KeyMatcher != nil {
		switch {
		case len(m.rule.KeyMatcher.Name) > 0:
			name := m.rule.KeyMatcher.Name
			keyMatched = keyPath.HasMatchingSuffix(ctlres.Path{{MapKey: &name}})

		case len(m.rule.KeyMatcher.Path) > 0:
			keyMatched = m.rule.KeyMatcher.Path.Matches(keyPath)

		default:
			panic("Unknown search rule key matcher")
		}
	} else {
		keyMatched = true
	}

	if m.rule.ValueMatcher != nil {
		switch {
		case len(m.rule.ValueMatcher.Image) > 0:
			if reflect.DeepEqual(m.rule.ValueMatcher.Image, value) {
				valueMatched = true
			}

		case len(m.rule.ValueMatcher.ImageRepo) > 0:
			if valueStr, ok := value.(string); ok {
				repo, matchesImg := ctlimg.URLRepo(valueStr)
				if matchesImg && m.rule.ValueMatcher.ImageRepo == repo {
					valueMatched = true
				}
			}

		default:
			panic("Unknown search rule value matcher")
		}
	} else {
		valueMatched = true
	}

	return keyMatched && valueMatched, m.rule.UpdateStrategyWithDefaults()
}
