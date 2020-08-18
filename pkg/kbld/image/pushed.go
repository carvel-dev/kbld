// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

// PushedImage respresents an image that will be pushed when its URL is requested
type PushedImage struct {
	image  Image
	imgDst ctlconf.ImageDestination
	docker Docker
}

func NewPushedImage(image Image, imgDst ctlconf.ImageDestination, docker Docker) PushedImage {
	return PushedImage{image, imgDst, docker}
}

func (i PushedImage) URL() (string, []ImageMeta, error) {
	url, metas, err := i.image.URL()
	if err != nil {
		return "", nil, err
	}

	digest, err := i.docker.Push(DockerTmpRef{url}, i.imgDst.NewImage)
	if err != nil {
		return "", nil, err
	}

	url, metas2, err := NewDigestedImageFromParts(i.imgDst.NewImage, digest.AsString()).URL()

	metas = append(metas, metas2...)

	return url, metas, err
}
