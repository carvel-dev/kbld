package image

import (
	"path/filepath"

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

type BuiltImageSources struct {
	Git   *BuiltImageSourceGit   `json:",omitempty" yaml:",omitempty"`
	Local *BuiltImageSourceLocal `json:",omitempty" yaml:",omitempty"`
}

type BuiltImageSourceGit struct {
	RemoteURL string `json:",omitempty" yaml:",omitempty"`
	SHA       string
	Dirty     bool
	Tags      []string `json:",omitempty" yaml:",omitempty"`
}

type BuiltImageSourceLocal struct {
	Path string
}

func (i BuiltImage) Sources() (BuiltImageSources, error) {
	absPath, err := filepath.Abs(i.buildSource.Path)
	if err != nil {
		return BuiltImageSources{}, err
	}

	info := BuiltImageSources{
		Local: &BuiltImageSourceLocal{
			Path: absPath,
		},
	}

	gitRepo := NewGitRepo(absPath)

	if gitRepo.IsValid() {
		var err error
		info.Git = &BuiltImageSourceGit{}

		info.Git.RemoteURL, err = gitRepo.RemoteURL()
		if err != nil {
			return BuiltImageSources{}, err
		}

		info.Git.SHA, err = gitRepo.HeadSHA()
		if err != nil {
			return BuiltImageSources{}, err
		}

		info.Git.Dirty, err = gitRepo.IsDirty()
		if err != nil {
			return BuiltImageSources{}, err
		}

		info.Git.Tags, err = gitRepo.HeadTags()
		if err != nil {
			return BuiltImageSources{}, err
		}
	}

	return info, nil
}
