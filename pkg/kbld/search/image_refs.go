package search

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type ImageRefs struct {
	res         interface{}
	searchRules []ctlconf.SearchRule
}

type ImageRefsVisitorFunc func(interface{}) (interface{}, bool)

func NewImageRefs(res interface{}, searchRules []ctlconf.SearchRule) ImageRefs {
	return ImageRefs{res, searchRules}
}

func (refs ImageRefs) Visit(visitorFunc ImageRefsVisitorFunc) {
	visitorFunc.Apply(refs.res, refs.searchRules)
}

func (v ImageRefsVisitorFunc) Apply(res interface{}, searchRules []ctlconf.SearchRule) {
	tmpRefs := map[string]interface{}{}
	tmpRefPrefix := v.randomPrefix()
	tmpRefIdx := 0

	insertTmpRefsFunc := func(val interface{}) (interface{}, bool) {
		newVal, updated := v(val)
		if !updated {
			return nil, false
		}
		// Only use temporary references for values that are updated;
		// otherwise future search rules may find tmp refs
		tmpRefId := fmt.Sprintf(tmpRefPrefix+"%d__", tmpRefIdx)
		tmpRefs[tmpRefId] = newVal
		tmpRefIdx += 1
		return tmpRefId, true
	}

	// Use a single matcher that represents all rules instead
	// so that each leaf value (string) is found once
	// even if it matches multiple search rules
	NewFields(res, RulesMatcher{searchRules}).Visit(v.extractValueFunc(insertTmpRefsFunc))

	resolveTmpRefsFunc := func(val interface{}) (interface{}, bool) {
		if valStr, ok := val.(string); ok {
			if actualRef, found := tmpRefs[valStr]; found {
				delete(tmpRefs, valStr)
				return actualRef, true
			}
		}
		return nil, false
	}

	NewFields(res, tmpRefMatcher{tmpRefPrefix}).Visit(v.extractValueFunc(resolveTmpRefsFunc))

	if len(tmpRefs) > 0 {
		panic("ImageRefs: Expected all tmp refs to be found")
	}
}

func (v ImageRefsVisitorFunc) extractValueFunc(visitorFunc ImageRefsVisitorFunc) FieldsVisitorFunc {
	return func(val interface{}, ext ctlconf.SearchRuleUpdateStrategy) (interface{}, bool) {
		switch {
		case ext.EntireValue != nil:
			return visitorFunc(val)

		case ext.JSON != nil:
			return v.extractValueAsJSONorYAML(val, ext.JSON.SearchRules, false)

		case ext.YAML != nil:
			return v.extractValueAsJSONorYAML(val, ext.YAML.SearchRules, true)

		default:
			panic("Unknown extraction type")
		}
	}
}

func (v ImageRefsVisitorFunc) extractValueAsJSONorYAML(val interface{},
	searchRules []ctlconf.SearchRule, allowYAML bool) (interface{}, bool) {

	valStr, ok := val.(string)
	if !ok {
		return val, false
	}

	var decodedVal, decodedJSONVal, decodedYAMLVal interface{}
	var decodedAsYAML bool

	jsonErr := json.Unmarshal([]byte(valStr), &decodedJSONVal)
	yamlErr := yaml.Unmarshal([]byte(valStr), &decodedYAMLVal)
	switch {
	case jsonErr == nil:
		decodedVal = decodedJSONVal
	case allowYAML && jsonErr != nil && yamlErr == nil:
		decodedVal = decodedYAMLVal
		decodedAsYAML = true
	default:
		return val, false
	}

	v.Apply(decodedVal, searchRules)

	if decodedAsYAML {
		valBs, err := yaml.Marshal(decodedVal)
		if err != nil {
			panic(fmt.Sprintf("ObjVisitor: Encoding as YAML: %s", err))
		}
		return string(valBs), true
	}

	valBs, err := json.Marshal(decodedVal)
	if err != nil {
		panic(fmt.Sprintf("ObjVisitor: Encoding as JSON: %s", err))
	}

	return string(valBs), true
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

func (m tmpRefMatcher) Matches(key string, value interface{}) (bool, ctlconf.SearchRuleUpdateStrategy) {
	if valStr, ok := value.(string); ok {
		return strings.HasPrefix(valStr, m.prefix), (ctlconf.SearchRule{}).UpdateStrategyWithDefaults()
	}
	return false, ctlconf.SearchRuleUpdateStrategy{}
}
