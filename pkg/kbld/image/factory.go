package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type Image interface {
	URL() (string, error)
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
	if overrideURL, found := f.shouldOverride(url); found {
		url = overrideURL.NewImage
	}

	if buildSource, found := f.shouldBuild(url); found {
		docker := Docker{f.logger}
		buildImg := NewBuiltImage(url, buildSource, docker)

		if imgDst, found := f.shouldPush(url); found {
			return NewPushedImage(buildImg, imgDst, docker)
		}
		return buildImg
	}

	digestedImage := MaybeNewDigestedImage(url)
	if digestedImage != nil {
		return digestedImage
	}

	return NewResolvedImage(url, f.registry)
}

func (f Factory) NewBuilt(url string) (BuiltImage, bool) {
	// TODO consolidate somehow with New?
	if overrideURL, found := f.shouldOverride(url); found {
		url = overrideURL.NewImage
	}

	if buildSource, found := f.shouldBuild(url); found {
		docker := Docker{f.logger}
		return NewBuiltImage(url, buildSource, docker), true
	}

	return BuiltImage{}, false
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
