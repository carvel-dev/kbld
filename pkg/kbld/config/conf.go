package config

import (
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
	defaultRule := SearchRule{
		KeyMatcher: &SearchRuleKeyMatcher{Name: "image"},
	}
	return append([]SearchRule{defaultRule}, c.SearchRulesWithoutDefaults()...)
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
	return result
}
