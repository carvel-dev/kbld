// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package search

import (
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlres "carvel.dev/kbld/pkg/kbld/resources"
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
	f.visit(ctlres.Path{}, f.res, visitorFunc)
}

func (f Fields) visit(keyPath ctlres.Path, res interface{}, visitorFunc FieldsVisitorFunc) {
	switch typedObj := res.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			k := k // copy
			newKeyPath := append(f.newPath(keyPath), &ctlres.PathPart{MapKey: &k})

			if matched, ext := f.matcher.Matches(newKeyPath, v); matched {
				if newVal, update := visitorFunc(v, ext); update {
					typedObj[k] = newVal
				} else {
					f.visit(newKeyPath, typedObj[k], visitorFunc)
				}
			} else {
				f.visit(newKeyPath, typedObj[k], visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			k := k // copy
			newKeyPath := append(f.newPath(keyPath), &ctlres.PathPart{MapKey: &k})

			if matched, ext := f.matcher.Matches(newKeyPath, v); matched {
				if newVal, update := visitorFunc(v, ext); update {
					typedObj[k] = newVal.(string)
				} else {
					f.visit(newKeyPath, typedObj[k], visitorFunc)
				}
			} else {
				f.visit(newKeyPath, typedObj[k], visitorFunc)
			}
		}

	case []interface{}:
		for i, o := range typedObj {
			i := i // copy
			newKeyPath := append(f.newPath(keyPath), &ctlres.PathPart{
				ArrayIndex: &ctlres.PathPartArrayIndex{Index: &i},
			})

			if matched, ext := f.matcher.Matches(newKeyPath, o); matched {
				if newVal, update := visitorFunc(o, ext); update {
					typedObj[i] = newVal
				} else {
					f.visit(newKeyPath, o, visitorFunc)
				}
			} else {
				f.visit(newKeyPath, o, visitorFunc)
			}
		}
	}
}

func (Fields) newPath(p ctlres.Path) ctlres.Path {
	return ctlres.Path(append(ctlres.Path{}, p...))
}
