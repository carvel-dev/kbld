// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"path/filepath"

	ctlbdk "github.com/k14s/kbld/pkg/kbld/builder/docker"
	ctlbkb "github.com/k14s/kbld/pkg/kbld/builder/kubectlbuildkit"
	ctlbpk "github.com/k14s/kbld/pkg/kbld/builder/pack"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type BuiltImage struct {
	url         string
	buildSource ctlconf.Source
	imgDst      *ctlconf.ImageDestination

	docker          ctlbdk.Docker
	pack            ctlbpk.Pack
	kubectlBuildkit ctlbkb.KubectlBuildkit
}

func NewBuiltImage(url string, buildSource ctlconf.Source, imgDst *ctlconf.ImageDestination,
	docker ctlbdk.Docker, pack ctlbpk.Pack, kubectlBuildkit ctlbkb.KubectlBuildkit) BuiltImage {

	return BuiltImage{url, buildSource, imgDst, docker, pack, kubectlBuildkit}
}

func (i BuiltImage) URL() (string, []ImageMeta, error) {
	metas, err := i.sources()
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

		return i.optionalPushWithDocker(dockerTmpRef, metas)

	case i.buildSource.KubectlBuildkit != nil:
		url, err := i.kubectlBuildkit.BuildAndPush(
			urlRepo, i.buildSource.Path, i.imgDst, *i.buildSource.KubectlBuildkit)
		return url, metas, err

	default:
		if i.buildSource.Docker == nil {
			i.buildSource.Docker = &ctlconf.SourceDockerOpts{}
		}

		opts := ctlbdk.DockerBuildOpts{
			Target:     i.buildSource.Docker.Build.Target,
			Pull:       i.buildSource.Docker.Build.Pull,
			NoCache:    i.buildSource.Docker.Build.NoCache,
			File:       i.buildSource.Docker.Build.File,
			RawOptions: i.buildSource.Docker.Build.RawOptions,
		}

		dockerTmpRef, err := i.docker.Build(urlRepo, i.buildSource.Path, opts)
		if err != nil {
			return "", nil, err
		}

		return i.optionalPushWithDocker(dockerTmpRef, metas)
	}
}

func (i BuiltImage) optionalPushWithDocker(dockerTmpRef ctlbdk.DockerTmpRef, metas []ImageMeta) (string, []ImageMeta, error) {
	if i.imgDst != nil {
		digest, err := i.docker.Push(dockerTmpRef, i.imgDst.NewImage)
		if err != nil {
			return "", nil, err
		}

		url, metas2, err := NewDigestedImageFromParts(i.imgDst.NewImage, digest.AsString()).URL()
		if err != nil {
			return "", nil, err
		}

		return url, append(metas, metas2...), nil
	}

	return dockerTmpRef.AsString(), metas, nil
}

type BuiltImageSourceGit struct {
	Type      string // always set to 'git'
	RemoteURL string `json:",omitempty" yaml:",omitempty"`
	SHA       string
	Dirty     bool
	Tags      []string `json:",omitempty" yaml:",omitempty"`
}

func (BuiltImageSourceGit) meta() {}

type BuiltImageSourceLocal struct {
	Type string // always set to 'local'
	Path string
}

func (BuiltImageSourceLocal) meta() {}

func (i BuiltImage) sources() ([]ImageMeta, error) {
	var sources []ImageMeta

	absPath, err := filepath.Abs(i.buildSource.Path)
	if err != nil {
		return nil, err
	}

	sources = append(sources, BuiltImageSourceLocal{
		Type: "local",
		Path: absPath,
	})

	gitRepo := NewGitRepo(absPath)

	if gitRepo.IsValid() {
		var err error
		git := BuiltImageSourceGit{Type: "git"}

		git.RemoteURL, err = gitRepo.RemoteURL()
		if err != nil {
			return nil, err
		}

		git.SHA, err = gitRepo.HeadSHA()
		if err != nil {
			return nil, err
		}

		git.Dirty, err = gitRepo.IsDirty()
		if err != nil {
			return nil, err
		}

		git.Tags, err = gitRepo.HeadTags()
		if err != nil {
			return nil, err
		}

		sources = append(sources, git)
	}

	return sources, nil
}
