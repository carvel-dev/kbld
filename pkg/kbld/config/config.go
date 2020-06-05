package config

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/ghodss/yaml"
	semver "github.com/hashicorp/go-version"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/k14s/kbld/pkg/kbld/version"
)

const (
	configAPIVersion      = "kbld.k14s.io/v1alpha1"
	configKind            = "Config"
	sourcesKind           = "Sources"           // specify list of sources for building images
	imageOverridesKind    = "ImageOverrides"    // specify alternative image urls
	imageDestinationsKind = "ImageDestinations" // specify image push destinations
	imageKeysKind         = "ImageKeys"
)

var (
	configKinds = []string{configKind, sourcesKind, imageOverridesKind, imageDestinationsKind, imageKeysKind}
)

type Config struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind,omitempty"`

	MinimumRequiredVersion string `json:"minimumRequiredVersion,omitempty"`

	Sources      []Source           `json:"sources,omitempty"`
	Overrides    []ImageOverride    `json:"overrides,omitempty"`
	Destinations []ImageDestination `json:"destinations,omitempty"`
	Keys         []string           `json:"keys,omitempty"`
	SearchRules  []SearchRule       `json:"searchRules,omitempty"`
}

type Source struct {
	ImageRef
	Path string

	Docker *SourceDockerOpts
	Pack   *SourcePackOpts
}

type ImageOverride struct {
	ImageRef
	NewImage    string `json:"newImage"`
	Preresolved bool   `json:"preresolved,omitempty"`
	Metadata    `json:"metadata"`
}

type Metadata struct {
	URL        string        `json:"url"`
	SourceURLs []interface{} `json:"source_urls"`
}

type ImageDestination struct {
	ImageRef
	NewImage string `json:"newImage"`
}

type SearchRule struct {
	KeyMatcher     *SearchRuleKeyMatcher     `json:"keyMatcher,omitempty"`
	ValueMatcher   *SearchRuleValueMatcher   `json:"valueMatcher,omitempty"`
	UpdateStrategy *SearchRuleUpdateStrategy `json:"updateStrategy,omitempty"`
	// TODO ResourceMatchers (see kapp's matchers)
}

type SearchRuleKeyMatcher struct {
	Name string      `json:"name,omitempty"`
	Path ctlres.Path `json:"path,omitempty"`
	// TODO JSONPath string
}

type SearchRuleValueMatcher struct {
	Image     string `json:"image,omitempty"`
	ImageRepo string `json:"imageRepo,omitempty"`
	// TODO Regexp    string `json:"regexp,omitempty"`
}

type SearchRuleUpdateStrategy struct {
	None         *SearchRuleUpdateStrategyNone         `json:"none,omitempty"`
	EntireString *SearchRuleUpdateStrategyEntireString `json:"entireValue,omitempty"`
	JSON         *SearchRuleUpdateStrategyJSON         `json:"json,omitempty"`
	YAML         *SearchRuleUpdateStrategyYAML         `json:"yaml,omitempty"`
}

type SearchRuleUpdateStrategyNone struct{}

type SearchRuleUpdateStrategyEntireString struct{}

type SearchRuleUpdateStrategyJSON struct {
	SearchRules []SearchRule `json:"searchRules,omitempty"`
}

type SearchRuleUpdateStrategyYAML struct {
	SearchRules []SearchRule `json:"searchRules,omitempty"`
}

type ImageRef struct {
	Image     string `json:"image,omitempty"`
	ImageRepo string `json:"imageRepo,omitempty"`
}

func NewConfig() Config {
	return Config{
		APIVersion: configAPIVersion,
		Kind:       configKind,
	}
}

func NewConfigFromResource(res ctlres.Resource) (Config, error) {
	bs, err := res.AsYAMLBytes()
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Unmarshaling %s: %s", res.Description(), err)
	}

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("Validating %s: %s", res.Description(), err)
	}

	for i, imageDst := range config.Destinations {
		if len(imageDst.NewImage) == 0 {
			imageDst.NewImage = imageDst.Image
		}
		config.Destinations[i] = imageDst
	}

	return config, nil
}

func (d Config) Validate() error {
	if len(d.MinimumRequiredVersion) > 0 {
		if d.MinimumRequiredVersion[0] == 'v' {
			return fmt.Errorf("Validating minimum version: Must not have prefix 'v' (e.g. '0.8.0')")
		}

		userConstraint, err := semver.NewConstraint(">=" + d.MinimumRequiredVersion)
		if err != nil {
			return fmt.Errorf("Parsing minimum version constraint: %s", err)
		}

		kbldVersion, err := semver.NewVersion(version.Version)
		if err != nil {
			return fmt.Errorf("Parsing version constraint: %s", err)
		}

		if !userConstraint.Check(kbldVersion) {
			return fmt.Errorf("kbld version '%s' does "+
				"not meet the minimum required version '%s'", version.Version, d.MinimumRequiredVersion)
		}
	}

	for i, src := range d.Sources {
		err := src.Validate()
		if err != nil {
			return fmt.Errorf("Validating Sources[%d]: %s", i, err)
		}
	}

	for i, override := range d.Overrides {
		err := override.Validate()
		if err != nil {
			return fmt.Errorf("Validating Overrides[%d]: %s", i, err)
		}
	}

	for i, dst := range d.Destinations {
		err := dst.Validate()
		if err != nil {
			return fmt.Errorf("Validating Destinations[%d]: %s", i, err)
		}
	}

	for i, key := range d.Keys {
		if len(key) == 0 {
			return fmt.Errorf("Validating Destinations[%d]: Expected to be non-empty", i)
		}
	}

	for i, sr := range d.SearchRules {
		err := sr.Validate()
		if err != nil {
			return fmt.Errorf("Validating SearchRules[%d]: %s", i, err)
		}
	}

	return nil
}

func (d Source) Validate() error {
	err := d.ImageRef.Validate()
	if err != nil {
		return err
	}
	if len(d.Path) == 0 {
		return fmt.Errorf("Expected Path to be non-empty")
	}
	return nil
}

func (d ImageOverride) Validate() error {
	err := d.ImageRef.Validate()
	if err != nil {
		return err
	}
	if len(d.NewImage) == 0 {
		return fmt.Errorf("Expected NewImage to be non-empty")
	}
	return nil
}

func (d ImageDestination) Validate() error {
	return d.ImageRef.Validate()
}

func (d SearchRule) Validate() error {
	if d.KeyMatcher == nil && d.ValueMatcher == nil {
		return fmt.Errorf("Expected KeyMatcher or ValueMatcher to be non-empty")
	}
	if d.KeyMatcher != nil {
		if len(d.KeyMatcher.Name) == 0 && len(d.KeyMatcher.Path) == 0 {
			return fmt.Errorf("Expected KeyMatcher.Name or KeyMatcher.Path to be non-empty")
		}
	}
	if d.ValueMatcher != nil {
		if len(d.ValueMatcher.Image) == 0 && len(d.ValueMatcher.ImageRepo) == 0 {
			return fmt.Errorf("Expected ValueMatcher.Image or ValueMatcher.ImageRepo to be non-empty")
		}
	}
	return nil
}

func (r ImageRef) Validate() error {
	if len(r.Image) == 0 && len(r.ImageRepo) == 0 {
		return fmt.Errorf("Expected Image or ImageRepo to be non-empty")
	}
	return nil
}

func (d Config) AsBytes() ([]byte, error) {
	bs, err := yaml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("Marshaling config: %s", err)
	}

	return bs, nil
}

func (d Config) WriteToFile(path string) error {
	bs, err := d.AsBytes()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bs, 0700)
	if err != nil {
		return fmt.Errorf("Writing lock config: %s", err)
	}

	return nil
}

func UniqueImageOverrides(overrides []ImageOverride) []ImageOverride {
	var result []ImageOverride
	for _, override := range overrides {
		var found bool
		for _, addedOverride := range result {
			if reflect.DeepEqual(addedOverride, override) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, override)
		}
	}
	return result
}

func (r SearchRule) UpdateStrategyWithDefaults() SearchRuleUpdateStrategy {
	if r.UpdateStrategy != nil {
		return *r.UpdateStrategy
	}
	return SearchRuleUpdateStrategy{
		EntireString: &SearchRuleUpdateStrategyEntireString{},
	}
}
