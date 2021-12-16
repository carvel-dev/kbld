// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
)

// TaggedImage represents an image that will be tagged when its URL is requested
type TaggedImage struct {
	image    Image
	imgDst   ctlconf.ImageDestination
	registry ctlreg.Registry
}

func NewTaggedImage(image Image, imgDst ctlconf.ImageDestination, registry ctlreg.Registry) TaggedImage {
	return TaggedImage{image, imgDst, registry}
}

func (i TaggedImage) URL() (string, []ctlconf.Origin, error) {
	url, origins, err := i.image.URL()
	if err != nil {
		return "", nil, err
	}

	if len(i.imgDst.Tags) > 0 {
		dstRef, err := regname.NewDigest(url, regname.WeakValidation)
		if err != nil {
			return "", nil, err
		}

		// Keep this ref separate to avoid any kind of modification
		// when changing tag on the dst ref
		srcRef, err := regname.NewDigest(url, regname.WeakValidation)
		if err != nil {
			return "", nil, err
		}

		for _, tag := range i.imgDst.Tags {
			err := i.registry.WriteTag(dstRef.Context().Tag(tag), srcRef)
			if err != nil {
				return "", nil, err
			}
		}

		origins = append(origins, ctlconf.Origin{Tagged: &ctlconf.OriginTagged{Tags: i.imgDst.Tags}})
	}

	return url, origins, err
}
