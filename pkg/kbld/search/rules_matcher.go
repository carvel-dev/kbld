package search

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
)

type RulesMatcher struct {
	rules []ctlconf.SearchRule
}

func NewRulesMatcher(rules []ctlconf.SearchRule) RulesMatcher {
	return RulesMatcher{rules}
}

var _ Matcher = RuleMatcher{}

func (m RulesMatcher) Matches(keyPath ctlres.Path, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy) {
	for _, rule := range m.rules {
		matches, extraction := (RuleMatcher{rule}).Matches(keyPath, value)
		if matches {
			return true, extraction
		}
	}
	return false, ctlconf.SearchRuleUpdateStrategy{}
}
