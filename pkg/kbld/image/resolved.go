package image

import (
	regname "github.com/google/go-containerregistry/pkg/name"
)

// ResolvedImage respresents an image that will be resolved into url+digest
type ResolvedImage struct {
	url      string
	registry Registry
}

func NewResolvedImage(url string, registry Registry) ResolvedImage {
	return ResolvedImage{url, registry}
}

func (i ResolvedImage) URL() (string, error) {
	tag, err := regname.NewTag(i.url, regname.WeakValidation)
	if err != nil {
		return "", err
	}

	imgDescriptor, err := i.registry.Generic(tag)
	if err != nil {
		return "", err
	}

	return NewDigestedImageFromParts(tag.Repository.String(), imgDescriptor.Digest.String()).URL()
}
