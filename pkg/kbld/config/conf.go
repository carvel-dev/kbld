// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"reflect"

	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
)

type Conf struct {
	configs []Config
}

func NewConfFromResources(resources []ctlres.Resource) ([]ctlres.Resource, Conf, error) {
	var rsWithoutConfigs []ctlres.Resource
	var configs []Config

	for _, res := range resources {
		if res.APIVersion() == configAPIVersion && matchesConfigKind(res) {
			config, err := NewConfigFromResource(res)
			if err != nil {
				return nil, Conf{}, err
			}
			configs = append(configs, config)
		} else {
			rsWithoutConfigs = append(rsWithoutConfigs, res)
		}
	}

	return rsWithoutConfigs, Conf{configs}, nil
}

func (c Conf) WithAdditionalConfig(config Config) Conf {
	newConf := Conf{}
	newConf.configs = append([]Config{}, newConf.configs...)
	newConf.configs = append(newConf.configs, config)
	return newConf
}

func matchesConfigKind(res ctlres.Resource) bool {
	for _, kind := range configKinds {
		if res.Kind() == kind {
			return true
		}
	}
	return false
}

func (c Conf) Sources() []Source {
	var result []Source
	for _, config := range c.configs {
		result = append(result, config.Sources...)
	}
	return result
}

func (c Conf) ImageOverrides() []ImageOverride {
	var result []ImageOverride
	for _, config := range c.configs {
		result = append(result, config.Overrides...)
	}
	return result
}

func (c Conf) ImageDestinations() []ImageDestination {
	var result []ImageDestination
	for _, config := range c.configs {
		result = append(result, config.Destinations...)
	}
	return result
}

func (c Conf) SearchRules() []SearchRule {
	result := append([]SearchRule{}, c.SearchRulesWithoutDefaults()...)

	// Add default image rule at the end so that
	// there is an opportunity to match image kv with other rules
	result = append(result, SearchRule{
		KeyMatcher: &SearchRuleKeyMatcher{Name: "image"},
	})

	return c.dedupSearchRules(result)
}

func (c Conf) SearchRulesWithoutDefaults() []SearchRule {
	result := []SearchRule{}
	for _, config := range c.configs {
		for _, key := range config.Keys {
			result = append(result, SearchRule{
				KeyMatcher: &SearchRuleKeyMatcher{Name: key},
			})
		}
	}
	for _, config := range c.configs {
		result = append(result, config.SearchRules...)
	}
	return c.dedupSearchRules(result)
}

func (c Conf) dedupSearchRules(rules []SearchRule) []SearchRule {
	var result []SearchRule
	for _, rule := range rules {
		var alreadySaved bool
		for _, savedRule := range result {
			if reflect.DeepEqual(rule, savedRule) {
				alreadySaved = true
				break
			}
		}
		if !alreadySaved {
			result = append(result, rule)
		}
	}
	return result
}
