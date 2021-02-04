// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlbdk "github.com/k14s/kbld/pkg/kbld/builder/docker"
	ctlbko "github.com/k14s/kbld/pkg/kbld/builder/ko"
	ctlbkb "github.com/k14s/kbld/pkg/kbld/builder/kubectlbuildkit"
	ctlbpk "github.com/k14s/kbld/pkg/kbld/builder/pack"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctllog "github.com/k14s/kbld/pkg/kbld/logger"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
)

type Image interface {
	URL() (string, []Meta, error)
}

type Meta interface {
	meta()
}

type Factory struct {
	conf     ctlconf.Conf
	registry ctlreg.Registry
	logger   ctllog.Logger
}

func NewFactory(conf ctlconf.Conf, registry ctlreg.Registry, logger ctllog.Logger) Factory {
	return Factory{conf, registry, logger}
}

func (f Factory) New(url string) Image {
	if overrideConf, found := f.shouldOverride(url); found {
		url = overrideConf.NewImage
		if overrideConf.Preresolved {
			return NewPreresolvedImage(url)
		} else if overrideConf.TagSelection != nil {
			return NewTagSelectedImage(url, overrideConf.TagSelection, f.registry)
		}
	}

	if srcConf, found := f.shouldBuild(url); found {
		imgDstConf := f.optionalPushConf(url)

		docker := ctlbdk.NewDocker(f.logger)
		pack := ctlbpk.NewPack(docker, f.logger)
		kubectlBuildkit := ctlbkb.NewKubectlBuildkit(f.logger)
		ko := ctlbko.NewKo(f.logger)

		builtImg := NewBuiltImage(url, srcConf, imgDstConf,
			docker, pack, kubectlBuildkit, ko)

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
