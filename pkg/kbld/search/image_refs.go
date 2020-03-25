package search

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type ImageRefs struct {
	resource    interface{}
	searchRules []ctlconf.SearchRule
}

func NewImageRefs(res interface{}, searchRules []ctlconf.SearchRule) ImageRefs {
	return ImageRefs{res, searchRules}
}

func (kvs ImageRefs) Visit(visitorFunc func(interface{}) (interface{}, bool)) {
	for _, sr := range kvs.searchRules {
		kvs.visitValues(kvs.resource, Matcher{sr}, visitorFunc)
	}
}

func (kvs ImageRefs) visitValues(obj interface{}, matcher Matcher, visitorFunc func(interface{}) (interface{}, bool)) {
	switch typedObj := obj.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if matcher.Matches(k, v) {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal
				}
			} else {
				kvs.visitValues(typedObj[k], matcher, visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			if matcher.Matches(k, v) {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal.(string)
				}
			} else {
				kvs.visitValues(typedObj[k], matcher, visitorFunc)
			}
		}

	case []interface{}:
		for _, o := range typedObj {
			kvs.visitValues(o, matcher, visitorFunc)
		}
	}
}
