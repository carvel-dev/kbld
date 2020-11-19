// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
)

// TagSelectedImage respresents an image that will be resolved into url+digest
type TagSelectedImage struct {
	url       string
	selection *versions.VersionSelection
	registry  ctlreg.Registry
}

func NewTagSelectedImage(url string, selection *versions.VersionSelection,
	registry ctlreg.Registry) TagSelectedImage {

	return TagSelectedImage{url, selection, registry}
}

func (i TagSelectedImage) URL() (string, []ImageMeta, error) {
	repo, err := regname.NewRepository(i.url, regname.WeakValidation)
	if err != nil {
		return "", nil, err
	}

	var tag string

	switch {
	case i.selection.Semver != nil:
		tags, err := i.registry.ListTags(repo)
		if err != nil {
			return "", nil, err
		}

		matchedVers := versions.NewSemvers(tags).FilterPrereleases(i.selection.Semver.Prereleases)

		if len(i.selection.Semver.Constraints) > 0 {
			matchedVers, err = matchedVers.FilterConstraints(i.selection.Semver.Constraints)
			if err != nil {
				return "", nil, fmt.Errorf("Selecting versions: %s", err)
			}
		}

		highestVersion, found := matchedVers.Highest()
		if !found {
			return "", nil, fmt.Errorf("Expected to find at least one version, but did not")
		}

		tag = highestVersion

	default:
		return "", nil, fmt.Errorf("Unknown tag selection strategy")
	}

	// tag value is included by ResolvedImage
	return NewResolvedImage(i.url+":"+tag, i.registry).URL()
}
