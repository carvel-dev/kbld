// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
)

// ResolvedImage respresents an image that will be resolved into url+digest
type ResolvedImage struct {
	url      string
	registry ctlreg.Registry
}

type ResolvedImageSourceURL struct {
	Type string // always set to 'resolved'
	URL  string
	Tag  string
}

func (ResolvedImageSourceURL) meta() {}

func NewResolvedImage(url string, registry ctlreg.Registry) ResolvedImage {
	return ResolvedImage{url, registry}
}

func (i ResolvedImage) URL() (string, []Meta, error) {
	tag, err := regname.NewTag(i.url, regname.WeakValidation)
	if err != nil {
		return "", nil, err
	}

	imgDescriptor, err := i.registry.Generic(tag)
	if err != nil {
		return "", nil, err
	}

	// Resolve image second time because some older registry can
	// return "random" digests that change for every request.
	// See https://github.com/vmware-tanzu/carvel-kbld/issues/21 for details.
	imgDescriptor2, err := i.registry.Generic(tag)
	if err != nil {
		return "", nil, err
	}

	if imgDescriptor.Digest.String() != imgDescriptor2.Digest.String() {
		return "", nil, fmt.Errorf("Expected digest resolution to be consistent over two separate requests")
	}

	url, metas, err := NewDigestedImageFromParts(tag.Repository.String(), imgDescriptor.Digest.String()).URL()
	if err != nil {
		return "", nil, err
	}

	metas = append(metas, ResolvedImageSourceURL{Type: "resolved", URL: i.url, Tag: tag.TagStr()})

	return url, metas, nil
}
