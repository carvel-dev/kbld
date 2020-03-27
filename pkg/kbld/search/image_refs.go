package search

import (
	"crypto/rand"
	"fmt"
	"strings"

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
	tmpRefs := map[string]interface{}{}
	tmpRefPrefix := kvs.randomPrefix()
	tmpRefIdx := 0

	insertTmpRefsFunc := func(val interface{}) (interface{}, bool) {
		newVal, updated := visitorFunc(val)
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
	NewObjVisitor(kvs.resource, RulesMatcher{kvs.searchRules}).Visit(insertTmpRefsFunc)

	resolveTmpRefsFunc := func(val interface{}) (interface{}, bool) {
		if valStr, ok := val.(string); ok {
			if actualRef, found := tmpRefs[valStr]; found {
				delete(tmpRefs, valStr)
				return actualRef, true
			}
		}
		return nil, false
	}

	NewObjVisitor(kvs.resource, tmpRefMatcher{tmpRefPrefix}).Visit(resolveTmpRefsFunc)

	if len(tmpRefs) > 0 {
		panic("ImageRefs: Expected all tmp refs to be found")
	}
}

func (kvs ImageRefs) randomPrefix() string {
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

func (m tmpRefMatcher) Matches(key string, value interface{}) bool {
	if valStr, ok := value.(string); ok {
		return strings.HasPrefix(valStr, m.prefix)
	}
	return false
}
