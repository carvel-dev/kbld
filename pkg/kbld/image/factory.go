// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"

	ctlbbz "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/bazel"
	ctlbdk "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/docker"
	ctlbko "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/ko"
	ctlbkb "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/kubectlbuildkit"
	ctlbpk "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/pack"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctllog "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/logger"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
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
	Conf           ctlconf.Conf
	AllowedToBuild bool
}

func NewFactory(opts FactoryOpts, registry ctlreg.Registry, logger ctllog.Logger) Factory {
	return Factory{opts, registry, logger}
}

func (f Factory) New(url string) Image {
	if overrideConf, found := f.shouldOverride(url); found {
		url = overrideConf.NewImage
		if overrideConf.Preresolved {
			return NewPreresolvedImage(url, overrideConf.ImageOrigins)
		} else if overrideConf.TagSelection != nil {
			return NewTagSelectedImage(url, overrideConf.TagSelection, f.registry)
		}
	}

	if srcConf, found := f.shouldBuild(url); found {
		if !f.opts.AllowedToBuild {
			return NewErrImage(fmt.Errorf("Building of images is disallowed (tried to build '%s' because a source was configured for it)", url))
		}

		imgDstConf := f.optionalPushConf(url)

		docker := ctlbdk.NewDocker(f.logger)
		dockerBuildx := ctlbdk.NewDockerBuildx(docker, f.logger)
		pack := ctlbpk.NewPack(docker, f.logger)
		kubectlBuildkit := ctlbkb.NewKubectlBuildkit(f.logger)
		ko := ctlbko.NewKo(f.logger)
		bazel := ctlbbz.NewBazel(docker, f.logger)

		builtImg := NewBuiltImage(url, srcConf, imgDstConf,
			docker, dockerBuildx, pack, kubectlBuildkit, ko, bazel)

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
