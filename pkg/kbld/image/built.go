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

func (i BuiltImage) URL() (string, []ImageMeta, error) {
	sources, err := i.sources()
	if err != nil {
		return "", nil, err
	}

	opts := DockerBuildOpts{
		Target:     i.buildSource.Docker.Build.Target,
		Pull:       i.buildSource.Docker.Build.Pull,
		NoCache:    i.buildSource.Docker.Build.NoCache,
		File:       i.buildSource.Docker.Build.File,
		RawOptions: i.buildSource.Docker.Build.RawOptions,
	}

	digest, err := i.docker.Build(i.url, i.buildSource.Path, opts)
	if err != nil {
		return "", nil, err
	}

	return digest.AsString(), sources, nil
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
