package search

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type Fields struct {
	res     interface{}
	matcher Matcher
}

type FieldsVisitorFunc func(interface{}, ctlconf.SearchRuleUpdateStrategy) (interface{}, bool)

func NewFields(res interface{}, matcher Matcher) Fields {
	return Fields{res, matcher}
}

func (f Fields) Visit(visitorFunc FieldsVisitorFunc) {
	f.visit(f.res, visitorFunc)
}

func (f Fields) visit(res interface{}, visitorFunc FieldsVisitorFunc) {
	switch typedObj := res.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if matched, ext := f.matcher.Matches(k, v); matched {
				if newVal, update := visitorFunc(v, ext); update {
					typedObj[k] = newVal
				}
			} else {
				f.visit(typedObj[k], visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			if matched, ext := f.matcher.Matches(k, v); matched {
				if newVal, update := visitorFunc(v, ext); update {
					typedObj[k] = newVal.(string)
				}
			} else {
				f.visit(typedObj[k], visitorFunc)
			}
		}

	case []interface{}:
		for _, o := range typedObj {
			f.visit(o, visitorFunc)
		}
	}
}
