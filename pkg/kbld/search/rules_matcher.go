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

func (m RulesMatcher) Matches(key string, value interface{}) bool {
	for _, rule := range m.rules {
		if (RuleMatcher{rule}).Matches(key, value) {
			return true
		}
	}
	return false
}
