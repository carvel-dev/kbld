// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package search

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"

	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlres "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/resources"
	"sigs.k8s.io/yaml"
)

type ImageRefs struct {
	res         interface{}
	searchRules []ctlconf.SearchRule
}

type ImageRefsVisitorFunc func(string) (string, bool)

func NewImageRefs(res interface{}, searchRules []ctlconf.SearchRule) ImageRefs {
	return ImageRefs{res, searchRules}
}

func (refs ImageRefs) Visit(visitorFunc ImageRefsVisitorFunc) {
	visitorFunc.Apply(refs.res, refs.searchRules)
}

func (v ImageRefsVisitorFunc) Apply(res interface{}, searchRules []ctlconf.SearchRule) {
	tmpRefs := map[string]string{}
	tmpRefPrefix := v.randomPrefix()
	tmpRefIdx := 0

	insertTmpRefsFunc := func(val string) (string, bool) {
		newVal, updated := v(val)
		if !updated {
			return "", false
		}
		// Only use temporary references for values that are updated;
		// otherwise future search rules may find tmp refs
		tmpRefID := fmt.Sprintf(tmpRefPrefix+"%d__", tmpRefIdx)
		tmpRefs[tmpRefID] = newVal
		tmpRefIdx++
		return tmpRefID, true
	}

	// Use a single matcher that represents all rules instead
	// so that each leaf value (string) is found once
	// even if it matches multiple search rules
	NewFields(res, RulesMatcher{searchRules}).Visit(v.extractValueFunc(insertTmpRefsFunc))

	resolveTmpRefsFunc := func(val string) (string, bool) {
		if actualRef, found := tmpRefs[val]; found {
			delete(tmpRefs, val)
			return actualRef, true
		}
		return "", false // TODO panic?
	}

	NewFields(res, tmpRefMatcher{tmpRefPrefix}).Visit(v.extractValueFunc(resolveTmpRefsFunc))

	if len(tmpRefs) > 0 {
		panic("ImageRefs: Expected all tmp refs to be found")
	}
}

func (v ImageRefsVisitorFunc) extractValueFunc(visitorFunc ImageRefsVisitorFunc) FieldsVisitorFunc {
	return func(val interface{}, ext ctlconf.SearchRuleUpdateStrategy) (interface{}, bool) {
		switch {
		case ext.None != nil:
			return val, false

		case ext.EntireString != nil:
			valStr, ok := val.(string)
			if !ok {
				return val, false
			}
			return visitorFunc(valStr)

		case ext.JSON != nil:
			return v.extractValueAsJSON(val, ext.JSON.SearchRules)

		case ext.YAML != nil:
			// Prefer to decode as JSON since JSON is valid YAML.
			// Only works for a single YAML document value.
			val, updated := v.extractValueAsJSON(val, ext.YAML.SearchRules)
			if updated {
				return val, updated
			}

			return v.extractValueAsYAML(val, ext.YAML.SearchRules)

		default:
			panic("Unknown extraction type")
		}
	}
}

func (v ImageRefsVisitorFunc) extractValueAsJSON(val interface{},
	searchRules []ctlconf.SearchRule) (interface{}, bool) {

	valStr, ok := val.(string)
	if !ok {
		return val, false
	}

	var decodedVal interface{}

	err := json.Unmarshal([]byte(valStr), &decodedVal)
	if err != nil {
		return val, false
	}

	v.Apply(decodedVal, searchRules)

	valBs, err := json.Marshal(decodedVal)
	if err != nil {
		panic(fmt.Sprintf("ObjVisitor: Encoding as JSON: %s", err))
	}

	return string(valBs), true
}

func (v ImageRefsVisitorFunc) extractValueAsYAML(val interface{},
	searchRules []ctlconf.SearchRule) (interface{}, bool) {

	valStr, ok := val.(string)
	if !ok {
		return val, false
	}

	docs, err := ctlres.NewYAMLFile(ctlres.NewBytesSource([]byte(valStr))).Docs()
	if err != nil {
		return val, false
	}

	var decodedVals []interface{}

	// Parse all documents before trying to search thru them
	for _, doc := range docs {
		var decodedVal interface{}

		err := yaml.Unmarshal(doc, &decodedVal)
		if err != nil {
			return val, false
		}

		// Skip over empty documents
		if decodedVal != nil {
			decodedVals = append(decodedVals, decodedVal)
		}
	}

	var result string

	for _, decodedVal := range decodedVals {
		v.Apply(decodedVal, searchRules)

		valBs, err := yaml.Marshal(decodedVal)
		if err != nil {
			panic(fmt.Sprintf("ObjVisitor: Encoding as YAML: %s", err))
		}

		result += "---\n" + string(valBs)
	}

	return result, true
}

func (ImageRefsVisitorFunc) randomPrefix() string {
	bs := make([]byte, 10)
	_, err := rand.Read(bs)
	if err != nil {
		panic("ImageRefs: Expected to fetch 10 random bytes")
	}
	var str string
	for _, b := range bs {
		str += fmt.Sprintf("%d", b)
	}
	return fmt.Sprintf("__kbld_image_ref_%s__", str)
}

type tmpRefMatcher struct {
	prefix string
}

var _ Matcher = tmpRefMatcher{}

func (m tmpRefMatcher) Matches(_ ctlres.Path, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy) {
	if valStr, ok := value.(string); ok {
		return strings.HasPrefix(valStr, m.prefix), (ctlconf.SearchRule{}).UpdateStrategyWithDefaults()
	}
	return false, ctlconf.SearchRuleUpdateStrategy{}
}
