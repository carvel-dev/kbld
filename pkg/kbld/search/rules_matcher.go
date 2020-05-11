package search

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type RulesMatcher struct {
	rules []ctlconf.SearchRule
}

func NewRulesMatcher(rules []ctlconf.SearchRule) RulesMatcher {
	return RulesMatcher{rules}
}

var _ Matcher = RuleMatcher{}

func (m RulesMatcher) Matches(key string, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy) {
	for _, rule := range m.rules {
		matches, extraction := (RuleMatcher{rule}).Matches(key, value)
		if matches {
			return true, extraction
		}
	}
	return false, ctlconf.SearchRuleUpdateStrategy{}
}
