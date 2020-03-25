package search_test

import (
	"reflect"
	"testing"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
)

func TestImageRefsMatches(t *testing.T) {
	type matcherExample struct {
		InputResource  interface{}
		OutputResource interface{}
		SearchRule     ctlconf.SearchRule
	}

	exs := []matcherExample{
		// By key
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "gcr.io/repo:something",
				},
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "found:gcr.io/repo:something",
				},
			},
			SearchRule: ctlconf.SearchRule{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
			},
		},
		// By exact image
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "gcr.io/repo:something",
				},
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "found:gcr.io/repo:something",
				},
			},
			SearchRule: ctlconf.SearchRule{
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					Image: "gcr.io/repo:something",
				},
			},
		},
		// By image repo
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "found:gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			SearchRule: ctlconf.SearchRule{
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					ImageRepo: "gcr.io/repo",
				},
			},
		},
		// By key and image repo
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key":   "gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"other": "gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key":   "found:gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"other": "gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			SearchRule: ctlconf.SearchRule{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					ImageRepo: "gcr.io/repo",
				},
			},
		},
	}

	for _, ex := range exs {
		refs := ctlser.NewImageRefs(ex.InputResource, []ctlconf.SearchRule{ex.SearchRule})
		refs.Visit(func(val interface{}) (interface{}, bool) { return "found:" + val.(string), true })
		if !reflect.DeepEqual(ex.InputResource, ex.OutputResource) {
			t.Fatalf("Expected %#v to succeed", ex)
		}
	}
}
