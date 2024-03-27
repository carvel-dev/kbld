// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"

	ctlbbz "carvel.dev/kbld/pkg/kbld/builder/bazel"
	ctlbdk "carvel.dev/kbld/pkg/kbld/builder/docker"
	ctlbko "carvel.dev/kbld/pkg/kbld/builder/ko"
	ctlbkb "carvel.dev/kbld/pkg/kbld/builder/kubectlbuildkit"
	ctlbpk "carvel.dev/kbld/pkg/kbld/builder/pack"
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
	ctlreg "carvel.dev/kbld/pkg/kbld/registry"
)

type Image interface {
	URL() (string, []ctlconf.Origin, error)
}

type Factory struct {
	opts     FactoryOpts
	registry ctlreg.Registry
	logger   ctllog.Logger
}

type FactoryOpts struct {
	Conf                    ctlconf.Conf
	AllowedToBuild          bool
	GlobalPlatformSelection *ctlconf.PlatformSelection
}

func NewFactory(opts FactoryOpts, registry ctlreg.Registry, logger ctllog.Logger) Factory {
	return Factory{opts, registry, logger}
}

func (f Factory) New(url string) Image {
	platformSelection := f.opts.GlobalPlatformSelection

	if overrideConf, found := f.shouldOverride(url); found {
		// Allow using same url but with additional selection (tag/platform)
		if len(overrideConf.NewImage) > 0 {
			url = overrideConf.NewImage
		}
		if overrideConf.PlatformSelection != nil {
			platformSelection = overrideConf.PlatformSelection
		}

		if overrideConf.Preresolved {
			// Do not support platform selection against explicitly configured image
			return NewPreresolvedImage(url, overrideConf.ImageOrigins)
		}
		if overrideConf.TagSelection != nil {
			tagSelected := NewTagSelectedImage(url, overrideConf.TagSelection, f.registry)
			return NewPlatformSelectedImage(tagSelected, platformSelection, f.registry)
		}
		// Continue on with potentially changed url or platform selection
	}

	if srcConf, found := f.shouldBuild(url); found {
		if !f.opts.AllowedToBuild {
			return NewErrImage(fmt.Errorf("Building of images is disallowed (tried to build '%s' because a source was configured for it)", url))
		}

		imgDstConf := f.optionalPushConf(url)

		docker := ctlbdk.New(f.logger)
		dockerBuildx := ctlbdk.NewBuildx(docker, f.logger)
		pack := ctlbpk.NewPack(docker, f.logger)
		kubectlBuildkit := ctlbkb.NewKubectlBuildkit(f.logger)
		ko := ctlbko.NewKo(f.logger)
		bazel := ctlbbz.NewBazel(docker, f.logger)

		var builtImg Image = NewBuiltImage(url, srcConf, imgDstConf,
			docker, dockerBuildx, pack, kubectlBuildkit, ko, bazel)

		if imgDstConf != nil {
			builtImg = NewTaggedImage(builtImg, *imgDstConf, f.registry)
		}
		return NewPlatformSelectedImage(builtImg, platformSelection, f.registry)
	}

	var resolvedImg Image
	if digestedImage := MaybeNewDigestedImage(url); digestedImage != nil {
		resolvedImg = digestedImage
	} else {
		resolvedImg = NewResolvedImage(url, f.registry)
	}
	return NewPlatformSelectedImage(resolvedImg, platformSelection, f.registry)
}

func (f Factory) shouldOverride(url string) (ctlconf.ImageOverride, bool) {
	urlMatcher := Matcher{url}
	for _, override := range f.opts.Conf.ImageOverrides() {
		if urlMatcher.Matches(override.ImageRef) {
			return override, true
		}
	}
	return ctlconf.ImageOverride{}, false
}

func (f Factory) shouldBuild(url string) (ctlconf.Source, bool) {
	urlMatcher := Matcher{url}
	for _, src := range f.opts.Conf.Sources() {
		if urlMatcher.Matches(src.ImageRef) {
			return src, true
		}
	}
	return ctlconf.Source{}, false
}

func (f Factory) optionalPushConf(url string) *ctlconf.ImageDestination {
	urlMatcher := Matcher{url}
	for _, dst := range f.opts.Conf.ImageDestinations() {
		if urlMatcher.Matches(dst.ImageRef) {
			return &dst
		}
	}
	return nil
}
