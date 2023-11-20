// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"path/filepath"

	ctlbbz "carvel.dev/kbld/pkg/kbld/builder/bazel"
	ctlbdk "carvel.dev/kbld/pkg/kbld/builder/docker"
	ctlbko "carvel.dev/kbld/pkg/kbld/builder/ko"
	ctlbkb "carvel.dev/kbld/pkg/kbld/builder/kubectlbuildkit"
	ctlbpk "carvel.dev/kbld/pkg/kbld/builder/pack"
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
)

type BuiltImage struct {
	url         string
	buildSource ctlconf.Source
	imgDst      *ctlconf.ImageDestination

	docker          ctlbdk.Docker
	dockerBuildx    ctlbdk.Buildx
	pack            ctlbpk.Pack
	kubectlBuildkit ctlbkb.KubectlBuildkit
	ko              ctlbko.Ko
	bazel           ctlbbz.Bazel
}

func NewBuiltImage(url string, buildSource ctlconf.Source, imgDst *ctlconf.ImageDestination,
	docker ctlbdk.Docker, dockerBuildx ctlbdk.Buildx, pack ctlbpk.Pack,
	kubectlBuildkit ctlbkb.KubectlBuildkit, ko ctlbko.Ko, bazel ctlbbz.Bazel) BuiltImage {

	return BuiltImage{url, buildSource, imgDst, docker, dockerBuildx, pack, kubectlBuildkit, ko, bazel}
}

func (i BuiltImage) URL() (string, []ctlconf.Origin, error) {
	origins, err := i.sources()
	if err != nil {
		return "", nil, err
	}

	urlRepo, _ := URLRepo(i.url)

	switch {
	case i.buildSource.Pack != nil:
		opts := ctlbpk.PackBuildOpts{
			Builder:    i.buildSource.Pack.Build.Builder,
			Buildpacks: i.buildSource.Pack.Build.Buildpacks,
			ClearCache: i.buildSource.Pack.Build.ClearCache,
			RawOptions: i.buildSource.Pack.Build.RawOptions,
		}

		dockerTmpRef, err := i.pack.Build(urlRepo, i.buildSource.Path, opts)
		if err != nil {
			return "", nil, err
		}

		return i.optionalPushWithDocker(dockerTmpRef, origins)

	case i.buildSource.KubectlBuildkit != nil:
		url, err := i.kubectlBuildkit.BuildAndPush(
			urlRepo, i.buildSource.Path, i.imgDst, *i.buildSource.KubectlBuildkit)
		return url, origins, err

	case i.buildSource.Ko != nil:
		dockerTmpRef, err := i.ko.Build(urlRepo, i.buildSource.Path, i.buildSource.Ko.Build)
		if err != nil {
			return "", nil, err
		}

		return i.optionalPushWithDocker(dockerTmpRef, origins)

	case i.buildSource.Bazel != nil:
		dockerTmpRef, err := i.bazel.Run(urlRepo, i.buildSource.Path, i.buildSource.Bazel.Run)
		if err != nil {
			return "", nil, err
		}

		return i.optionalPushWithDocker(dockerTmpRef, origins)

	case i.buildSource.Docker != nil && i.buildSource.Docker.Buildx != nil:
		url, err := i.dockerBuildx.BuildAndOptionallyPush(
			urlRepo, i.buildSource.Path, i.imgDst, *i.buildSource.Docker.Buildx)
		return url, origins, err

	// Fall back on Docker by default
	default:
		if i.buildSource.Docker == nil {
			i.buildSource.Docker = &ctlconf.SourceDockerOpts{}
		}

		opts := ctlbdk.BuildOpts{
			Target:     i.buildSource.Docker.Build.Target,
			Pull:       i.buildSource.Docker.Build.Pull,
			NoCache:    i.buildSource.Docker.Build.NoCache,
			File:       i.buildSource.Docker.Build.File,
			Buildkit:   i.buildSource.Docker.Build.Buildkit,
			RawOptions: i.buildSource.Docker.Build.RawOptions,
		}

		dockerTmpRef, err := i.docker.Build(urlRepo, i.buildSource.Path, opts)
		if err != nil {
			return "", nil, err
		}

		return i.optionalPushWithDocker(dockerTmpRef, origins)
	}
}

func (i BuiltImage) optionalPushWithDocker(dockerTmpRef ctlbdk.TmpRef, origins []ctlconf.Origin) (string, []ctlconf.Origin, error) {
	if i.imgDst != nil {
		digest, err := i.docker.Push(dockerTmpRef, i.imgDst.NewImage)
		if err != nil {
			return "", nil, err
		}

		url, moreOrigins, err := NewDigestedImageFromParts(i.imgDst.NewImage, digest.AsString()).URL()
		if err != nil {
			return "", nil, err
		}

		return url, append(origins, moreOrigins...), nil
	}

	return dockerTmpRef.AsString(), origins, nil
}

func (i BuiltImage) sources() ([]ctlconf.Origin, error) {
	var sources []ctlconf.Origin

	absPath, err := filepath.Abs(i.buildSource.Path)
	if err != nil {
		return nil, err
	}

	sources = append(sources, ctlconf.Origin{Local: &ctlconf.OriginLocal{Path: absPath}})

	gitRepo := NewGitRepo(absPath)

	if gitRepo.IsValid() {
		var err error

		sha, err := gitRepo.HeadSHA()
		if err != nil {
			return nil, err
		}

		unredactedURL, err := gitRepo.RemoteURL()
		if err != nil {
			return nil, err
		}

		git := ctlconf.OriginGit{
			SHA:       sha,
			RemoteURL: GitRedactedRemoteURL(unredactedURL),
		}

		git.Dirty, err = gitRepo.IsDirty()
		if err != nil {
			return nil, err
		}

		git.Tags, err = gitRepo.HeadTags()
		if err != nil {
			return nil, err
		}

		sources = append(sources, ctlconf.Origin{Git: &git})
	}

	return sources, nil
}
