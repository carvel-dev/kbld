// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
)

type Image interface {
	URL() (string, []ImageMeta, error)
}

type ImageMeta interface {
	meta()
}

type Factory struct {
	conf     ctlconf.Conf
	registry ctlreg.Registry
	logger   Logger
}

func NewFactory(conf ctlconf.Conf, registry ctlreg.Registry, logger Logger) Factory {
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
		imgDstConf := f.optionalPushConf(url)

		docker := Docker{f.logger}
		pack := Pack{docker, f.logger}
		kubectlBuildkit := KubectlBuildkit{f.logger}

		builtImg := NewBuiltImage(url, srcConf, imgDstConf,
			docker, pack, kubectlBuildkit)

		if imgDstConf != nil {
			return NewTaggedImage(builtImg, *imgDstConf, f.registry)
		}

		return builtImg
	}

	digestedImage := MaybeNewDigestedImage(url)
	if digestedImage != nil {
		return digestedImage
	}

	return NewResolvedImage(url, f.registry)
}

func (f Factory) shouldOverride(url string) (ctlconf.ImageOverride, bool) {
	urlMatcher := Matcher{url}
	for _, override := range f.conf.ImageOverrides() {
		if urlMatcher.Matches(override.ImageRef) {
			return override, true
		}
	}
	return ctlconf.ImageOverride{}, false
}

func (f Factory) shouldBuild(url string) (ctlconf.Source, bool) {
	urlMatcher := Matcher{url}
	for _, src := range f.conf.Sources() {
		if urlMatcher.Matches(src.ImageRef) {
			return src, true
		}
	}
	return ctlconf.Source{}, false
}

func (f Factory) optionalPushConf(url string) *ctlconf.ImageDestination {
	urlMatcher := Matcher{url}
	for _, dst := range f.conf.ImageDestinations() {
		if urlMatcher.Matches(dst.ImageRef) {
			return &dst
		}
	}
	return nil
}
