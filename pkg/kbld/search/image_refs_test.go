package search_test

import (
	"encoding/json"
	"reflect"
	"testing"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
)

func TestImageRefsMatches(t *testing.T) {
	type matcherExample struct {
		InputResource  interface{}
		OutputResource interface{}
		SearchRules    []ctlconf.SearchRule
		OutputImages   []string
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
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
			}},
			OutputImages: []string{"gcr.io/repo:something"},
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
			SearchRules: []ctlconf.SearchRule{{
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					Image: "gcr.io/repo:something",
				},
			}},
			OutputImages: []string{"gcr.io/repo:something"},
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
			SearchRules: []ctlconf.SearchRule{{
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					ImageRepo: "gcr.io/repo",
				},
			}},
			OutputImages: []string{"gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
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
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
				ValueMatcher: &ctlconf.SearchRuleValueMatcher{
					ImageRepo: "gcr.io/repo",
				},
			}},
			OutputImages: []string{"gcr.io/repo@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		},
		// Matchers matching same values
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
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
			}, {
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
			}},
			OutputImages: []string{"gcr.io/repo:something"},
		},
	}

	for _, ex := range exs {
		refs := ctlser.NewImageRefs(ex.InputResource, ex.SearchRules)

		foundImages := []string{}
		refs.Visit(func(val interface{}) (interface{}, bool) {
			foundImages = append(foundImages, val.(string))
			return "found:" + val.(string), true
		})

		if !reflect.DeepEqual(ex.InputResource, ex.OutputResource) {
			inBs, _ := json.Marshal(ex.InputResource)
			outBs, _ := json.Marshal(ex.OutputResource)
			t.Fatalf("Expected %#v to succeed: >>>%s<<< vs >>>%s<<<", ex, inBs, outBs)
		}
		if !reflect.DeepEqual(foundImages, ex.OutputImages) {
			t.Fatalf("Expected %#v to succeed: >>>%s<<< vs >>>%s<<<", ex, foundImages, ex.OutputImages)
		}
	}
}
