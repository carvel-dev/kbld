package search

import (
	"reflect"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

type Matcher struct {
	rule ctlconf.SearchRule
}

func (m Matcher) Matches(key string, value interface{}) bool {
	var keyMatched, valueMatched bool

	if m.rule.KeyMatcher != nil {
		if m.rule.KeyMatcher.Name == key {
			keyMatched = true
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

	return keyMatched && valueMatched
}
