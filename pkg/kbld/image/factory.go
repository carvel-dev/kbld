package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type Image interface {
	URL() (string, []ImageMeta, error)
}

type ImageMeta interface {
	meta()
}

type Factory struct {
	conf     ctlconf.Conf
	registry Registry
	logger   Logger
}

func NewFactory(conf ctlconf.Conf, registry Registry, logger Logger) Factory {
	return Factory{conf, registry, logger}
}

func (f Factory) New(url string) Image {
	if overrideConf, found := f.shouldOverride(url); found {
		url = overrideConf.NewImage
		if overrideConf.Preresolved {
			return NewPreresolvedImage(url)
		}
	}

	if srcConf, found := f.shouldBuild(url); found {
		docker := Docker{f.logger}
		buildImg := NewBuiltImage(url, srcConf, docker)

		if imgDstConf, found := f.shouldPush(url); found {
			return NewPushedImage(buildImg, imgDstConf, docker)
		}
		return buildImg
	}

	digestedImage := MaybeNewDigestedImage(url)
	if digestedImage != nil {
		return digestedImage
	}

	return NewResolvedImage(url, f.registry)
}

func (f Factory) shouldOverride(url string) (ctlconf.ImageOverride, bool) {
	for _, override := range f.conf.ImageOverrides() {
		if override.Image == url {
			return override, true
		}
	}
	return ctlconf.ImageOverride{}, false
}

func (f Factory) shouldBuild(url string) (ctlconf.Source, bool) {
	for _, source := range f.conf.Sources() {
		if source.Image == url {
			return source, true
		}
	}
	return ctlconf.Source{}, false
}

func (f Factory) shouldPush(url string) (ctlconf.ImageDestination, bool) {
	for _, dst := range f.conf.ImageDestinations() {
		if dst.Image == url {
			return dst, true
		}
	}
	return ctlconf.ImageDestination{}, false
}
