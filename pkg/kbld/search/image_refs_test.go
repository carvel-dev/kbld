// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package search_test

import (
	"encoding/json"
	"reflect"
	"sort"
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
		// JSON extraction
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"nginx1"}`,
				},
				"nestedimage": "nginx2",
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"found:nginx1"}`,
				},
				"nestedimage": "nginx2",
			},
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
				UpdateStrategy: &ctlconf.SearchRuleUpdateStrategy{
					JSON: &ctlconf.SearchRuleUpdateStrategyJSON{
						SearchRules: []ctlconf.SearchRule{{
							KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
								Name: "nestedimage",
							}},
						},
					},
				},
			}},
			OutputImages: []string{"nginx1"},
		},
		// YAML extraction
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"nginx1"}`,
				},
				"nestedimage": "nginx2",
				"key": `
nestedimage: nginx3
---
---
null
---
nestedimage: nginx4
`,
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"found:nginx1"}`,
				},
				"nestedimage": "nginx2",
				"key": `---
nestedimage: found:nginx3
---
nestedimage: found:nginx4
`,
			},
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
				UpdateStrategy: &ctlconf.SearchRuleUpdateStrategy{
					YAML: &ctlconf.SearchRuleUpdateStrategyYAML{
						SearchRules: []ctlconf.SearchRule{{
							KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
								Name: "nestedimage",
							}},
						},
					},
				},
			}},
			OutputImages: []string{"nginx1", "nginx3", "nginx4"},
		},
		// Nested YAML extraction
		{
			InputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"nginx1"}`,
				},
				"data": `
path:
- nestedimage: |
    something: nginx3
`,
			},
			OutputResource: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": `{"nestedimage":"found:nginx1"}`,
				},
				"data": `---
path:
- nestedimage: |
    ---
    something: found:nginx3
`,
			},
			SearchRules: []ctlconf.SearchRule{{
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "key",
				},
				UpdateStrategy: &ctlconf.SearchRuleUpdateStrategy{
					YAML: &ctlconf.SearchRuleUpdateStrategyYAML{
						SearchRules: []ctlconf.SearchRule{{
							KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
								Name: "nestedimage",
							}},
						},
					},
				},
			}, {
				KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
					Name: "data",
				},
				UpdateStrategy: &ctlconf.SearchRuleUpdateStrategy{
					YAML: &ctlconf.SearchRuleUpdateStrategyYAML{
						SearchRules: []ctlconf.SearchRule{{
							KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
								Name: "nestedimage",
							},
							UpdateStrategy: &ctlconf.SearchRuleUpdateStrategy{
								YAML: &ctlconf.SearchRuleUpdateStrategyYAML{
									SearchRules: []ctlconf.SearchRule{{
										KeyMatcher: &ctlconf.SearchRuleKeyMatcher{
											Name: "something",
										}},
									},
								},
							}},
						},
					},
				},
			}},
			OutputImages: []string{"nginx1", "nginx3"},
		},
		// Matching key within another matching key section
		{
			InputResource: map[string]interface{}{
				// key matches, but not expecting a map
				"key": map[string]interface{}{
					// nested-key matches as well and is a string
					"nested-key": "nginx1",
					"key":        "nginx2",
				},
				"another-key": map[string]interface{}{
					"key": "nginx3",
				},
			},
			OutputResource: map[string]interface{}{
				"key": map[string]interface{}{
					"nested-key": "found:nginx1",
					"key":        "found:nginx2",
				},
				"another-key": map[string]interface{}{
					"key": "found:nginx3",
				},
			},
			SearchRules: []ctlconf.SearchRule{
				{KeyMatcher: &ctlconf.SearchRuleKeyMatcher{Name: "key"}},
				{KeyMatcher: &ctlconf.SearchRuleKeyMatcher{Name: "nested-key"}},
			},
			OutputImages: []string{"nginx1", "nginx2", "nginx3"},
		},
	}

	for _, ex := range exs {
		refs := ctlser.NewImageRefs(ex.InputResource, ex.SearchRules)

		foundImages := []string{}
		refs.Visit(func(val string) (string, bool) {
			foundImages = append(foundImages, val)
			return "found:" + val, true
		})

		sort.Strings(foundImages)

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
