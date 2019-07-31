package image

import (
	regname "github.com/google/go-containerregistry/pkg/name"
)

type PreresolvedImage struct {
	url      string
	registry Registry
}

type PreresolvedImageSourceURL struct {
	Type string // always set to 'preresolved'
	URL  string
	Tag  string
}

func (PreresolvedImageSourceURL) meta() {}

func NewPreresolvedImage(url string, registry Registry) PreresolvedImage {
	return PreresolvedImage{url, registry}
}

func (i PreresolvedImage) URL() (string, []ImageMeta, error) {
	tag, err := regname.NewTag(i.url, regname.WeakValidation)
	if err != nil {
		return "", nil, err
	}

	imgDescriptor, err := i.registry.Generic(tag)
	if err != nil {
		return "", nil, err
	}

	_, metas, err := NewDigestedImageFromParts(tag.Repository.String(), imgDescriptor.Digest.String()).URL()
	if err != nil {
		return "", nil, err
	}

	metas = append(metas, PreresolvedImageSourceURL{Type: "preresolved", URL: i.url, Tag: tag.TagStr()})

	return i.url, metas, nil
}
