// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"io/ioutil"

	semver "github.com/hashicorp/go-version"
	"github.com/k14s/imgpkg/pkg/imgpkg/lockconfig"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/k14s/kbld/pkg/kbld/version"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"sigs.k8s.io/yaml"
)

const (
	configAPIVersion      = "kbld.k14s.io/v1alpha1"
	configKind            = "Config"
	sourcesKind           = "Sources"           // specify list of sources for building images
	imageOverridesKind    = "ImageOverrides"    // specify alternative image urls
	imageDestinationsKind = "ImageDestinations" // specify image push destinations
	imageKeysKind         = "ImageKeys"
)

type Kind struct {
	APIVersion, Kind string
}

var (
	configKinds = []Kind{
		{configAPIVersion, configKind},
		{configAPIVersion, configKind},
		{configAPIVersion, sourcesKind},
		{configAPIVersion, imageOverridesKind},
		{configAPIVersion, imageDestinationsKind},
		{configAPIVersion, imageKeysKind},
	}
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

	Docker          *SourceDockerOpts
	Pack            *SourcePackOpts
	KubectlBuildkit *SourceKubectlBuildkitOpts
	Ko              *SourceKoOpts
	Bazel           *SourceBazelOpts
}

type ImageOverride struct {
	ImageRef
	NewImage     string                     `json:"newImage"`
	Preresolved  bool                       `json:"preresolved,omitempty"`
	TagSelection *versions.VersionSelection `json:"tagSelection,omitempty"`
	ImageOrigins []Origin                   `json:"origins,omitempty"`
}

type ImageDestination struct {
	ImageRef
	NewImage string   `json:"newImage"`
	Tags     []string `json:"tags"`
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

func NewConfigFromImagesLock(res ctlres.Resource) (Config, error) {
	iLockBytes, err := res.AsYAMLBytes()
	if err != nil {
		return Config{}, err
	}

	imagesLock, err := lockconfig.NewImagesLockFromBytes(iLockBytes)
	if err != nil {
		return Config{}, fmt.Errorf("Unmarshaling %s as ImagesLock: %s", res.Description(), err)
	}

	overridesConfig := NewConfig()

	for _, image := range imagesLock.Images {
		imgOrigins, err := NewOriginsFromString(image.Annotations[ImagesLockKbldOrigins])
		if err != nil {
			return Config{}, fmt.Errorf("Unmarshaling %s as %s annotation:  %s", res.Description(), ImagesLockKbldOrigins, err)
		}
		iOverride := ImageOverride{
			ImageRef: ImageRef{
				Image: image.Annotations[ImagesLockKbldID],
			},
			NewImage:     image.Image,
			Preresolved:  true,
			ImageOrigins: imgOrigins,
		}
		overridesConfig.Overrides = append(overridesConfig.Overrides, iOverride)
	}

	err = overridesConfig.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("Validating %s: %s", res.Description(), err)
	}

	return overridesConfig, nil
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

	err = ioutil.WriteFile(path, bs, 0600)
	if err != nil {
		return fmt.Errorf("Writing lock config: %s", err)
	}

	return nil
}

// Equal reports whether this ImageOverride is equal to another ImageOverride.
//   (`ImageMeta` is descriptive — not identifying — so not part of equality)
func (d ImageOverride) Equal(other ImageOverride) bool {
	return d.ImageRef == other.ImageRef &&
		d.NewImage == other.NewImage &&
		d.Preresolved == other.Preresolved &&
		d.TagSelection == other.TagSelection
}

func UniqueImageOverrides(overrides []ImageOverride) []ImageOverride {
	var result []ImageOverride
	for _, override := range overrides {
		var found bool
		for _, addedOverride := range result {
			if override.Equal(addedOverride) {
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

func (d SearchRule) UpdateStrategyWithDefaults() SearchRuleUpdateStrategy {
	if d.UpdateStrategy != nil {
		return *d.UpdateStrategy
	}
	return SearchRuleUpdateStrategy{
		EntireString: &SearchRuleUpdateStrategyEntireString{},
	}
}
