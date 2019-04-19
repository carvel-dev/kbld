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

func (i PushedImage) URL() (string, error) {
	url, err := i.image.URL()
	if err != nil {
		return "", err
	}

	digest, err := i.docker.Push(DockerTmpRef{url}, i.imgDst.NewImage)
	if err != nil {
		return "", err
	}

	return NewDigestedImageFromParts(i.imgDst.NewImage, digest.AsString()).URL()
}
