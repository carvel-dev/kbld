package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type BuiltImage struct {
	url         string
	buildSource ctlconf.Source
	docker      Docker
}

func NewBuiltImage(url string, buildSource ctlconf.Source, docker Docker) BuiltImage {
	return BuiltImage{url, buildSource, docker}
}

func (i BuiltImage) URL() (string, error) {
	digest, err := i.docker.Build(i.url, i.buildSource.Path)
	if err != nil {
		return "", err
	}

	return digest.AsString(), nil
}
