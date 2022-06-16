// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	regtypes "github.com/google/go-containerregistry/pkg/v1/types"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
)

// PlatformSelectedImage selects specific image matching arch/platform
type PlatformSelectedImage struct {
	image     Image
	selection *ctlconf.PlatformSelection
	registry  ctlreg.Registry
}

func NewPlatformSelectedImage(image Image,
	selection *ctlconf.PlatformSelection, registry ctlreg.Registry) PlatformSelectedImage {

	return PlatformSelectedImage{image, selection, registry}
}

func (i PlatformSelectedImage) URL() (string, []ctlconf.Origin, error) {
	url, origins, err := i.image.URL()
	if err != nil {
		return url, origins, err
	}

	// Do nothing if platform selection is not requested
	if i.selection == nil {
		return url, origins, err
	}

	ref, err := regname.ParseReference(url)
	if err != nil {
		return "", nil, err
	}

	desc, err := i.registry.Generic(ref)
	if err != nil {
		return "", nil, err
	}

	switch desc.MediaType {
	case regtypes.OCIImageIndex, regtypes.DockerManifestList:
		imgIndex, err := i.registry.Index(ref)
		if err != nil {
			return "", nil, err
		}
		imgIndexManifest, err := imgIndex.IndexManifest()
		if err != nil {
			return "", nil, err
		}

		var matchedMan *regv1.Descriptor

		for _, man := range imgIndexManifest.Manifests {
			if man.Platform != nil && MatchesPlatformSelection(*man.Platform, *i.selection) {
				if matchedMan != nil {
					return "", nil, fmt.Errorf("Expected to find only one image under index '%s' with matching platform, but found more than one", url)
				}
				man := man // copy
				matchedMan = &man
			}
		}

		if matchedMan != nil {
			newURL, newOrigins, err := NewDigestedImageFromParts(ref.Context().Name(), matchedMan.Digest.String()).URL()
			if err != nil {
				return "", nil, err
			}
			newOrigins = append(newOrigins, ctlconf.Origin{
				PlatformSelected: &ctlconf.OriginPlatformSelected{
					Index:        url,
					OS:           i.selection.OS,
					Architecture: i.selection.Architecture,
					Variant:      i.selection.Variant,
				},
			})
			return newURL, append(origins, newOrigins...), nil
		}

		return "", nil, fmt.Errorf("Expected to find one image under index '%s' with matching platform, but found none", url)

	// Assume that if it's not an index, then image is all right to use
	default:
		return url, origins, nil
	}
}

// MatchesPlatformSelection checks if the given platform matches the required platforms.
// The given platform matches the required platform if
// - architecture and OS are identical.
// - OS version and variant are identical if provided.
// - features and OS features of the required platform are subsets of those of the given platform.
// Adapted from https://github.com/google/go-containerregistry/blob/570ba6c88a5041afebd4599981d849af96f5dba9/pkg/v1/remote/index.go#L263
func MatchesPlatformSelection(given regv1.Platform, required ctlconf.PlatformSelection) bool {
	// Required fields that must be identical.
	if given.Architecture != required.Architecture || given.OS != required.OS {
		return false
	}

	// Optional fields that may be empty, but must be identical if provided.
	if required.OSVersion != "" && given.OSVersion != required.OSVersion {
		return false
	}
	if required.Variant != "" && given.Variant != required.Variant {
		return false
	}

	// Verify required platform's features are a subset of given platform's features.
	if !isSubset(given.OSFeatures, required.OSFeatures) {
		return false
	}
	if !isSubset(given.Features, required.Features) {
		return false
	}

	return true
}

// isSubset checks if the required array of strings is a subset of the given lst.
func isSubset(lst, required []string) bool {
	set := make(map[string]bool)
	for _, value := range lst {
		set[value] = true
	}

	for _, value := range required {
		if _, ok := set[value]; !ok {
			return false
		}
	}

	return true
}
