// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"

	"sigs.k8s.io/yaml"
)

type Meta interface {
	meta()
}

type imageMeta struct {
	Metas []Meta `json:"metas,omitempty"`
}

const (
	GitMeta         = "git"
	LocalMeta       = "local"
	ResolvedMeta    = "resolved"
	TaggedMeta      = "tagged"
	PreresolvedMeta = "preresolved"
)

type BuiltImageSourceGit struct {
	Type      string   `json:"type"` // always set to GitMeta
	RemoteURL string   `json:"remoteUrl"`
	SHA       string   `json:"sha"`
	Dirty     bool     `json:"dirty"`
	Tags      []string `json:"tags,omitempty"`
}

type BuiltImageSourceLocal struct {
	Type string `json:"type"` // always set to LocalMeta
	Path string `json:"path"`
}

type ResolvedImageSourceURL struct {
	Type string `json:"type"` // always set to ResolvedMeta
	URL  string `json:"url"`
	Tag  string `json:"tag,omitempty"`
}

type TaggedImageMeta struct {
	Type string   `json:"type"` // always set to TaggedMeta
	Tags []string `json:"tags"`
}

type PreresolvedImageSourceURL struct {
	Type string `json:"type"` // always set to PreresolvedMeta
	URL  string `json:"url"`
	Tag  string `json:"tag,omitempty"`
}

func (BuiltImageSourceGit) meta()       {}
func (BuiltImageSourceLocal) meta()     {}
func (ResolvedImageSourceURL) meta()    {}
func (TaggedImageMeta) meta()           {}
func (PreresolvedImageSourceURL) meta() {}

func NewMetasFromString(metas string) ([]Meta, error) {
	imgMeta := imageMeta{}
	err := yaml.Unmarshal([]byte(metas), &imgMeta)
	if err != nil {
		return []Meta{}, err
	}
	return imgMeta.Metas, nil
}

var _ json.Unmarshaler = &imageMeta{}

func (m *imageMeta) UnmarshalJSON(data []byte) error {
	var list []interface{}
	err := yaml.Unmarshal(data, &list)
	if err != nil {
		return err
	}

	for _, item := range list {
		var local BuiltImageSourceLocal
		var git BuiltImageSourceGit
		var res ResolvedImageSourceURL
		var preres PreresolvedImageSourceURL
		var tag TaggedImageMeta

		yamlItem, _ := yaml.Marshal(&item)

		switch {
		case yaml.Unmarshal(yamlItem, &local) == nil && local.Type == LocalMeta:
			m.Metas = append(m.Metas, local)
		case yaml.Unmarshal(yamlItem, &git) == nil && git.Type == GitMeta:
			m.Metas = append(m.Metas, git)
		case yaml.Unmarshal(yamlItem, &res) == nil && res.Type == ResolvedMeta:
			m.Metas = append(m.Metas, res)
		case yaml.Unmarshal(yamlItem, &preres) == nil && preres.Type == PreresolvedMeta:
			m.Metas = append(m.Metas, preres)
		case yaml.Unmarshal(yamlItem, &tag) == nil && tag.Type == TaggedMeta:
			m.Metas = append(m.Metas, tag)
		default:
			// ignore unknown meta.
			// At this time...
			// - "Meta" are provided as primarily optional diagnostic information
			//   rather than operational data (read: less important). Losing
			//   this information does not change the correctness of kbld's
			//   primary purpose during deployment: to rewrite image references.
			//   It would be more than an annoyance to error-out if we were
			//   unable to parse such data.
			// - Ideally, yes, we'd at least report a warning. However, if there's
			//   a systemic condition (e.g. using an older version of kbld to
			//   deploy than was used to package) there would likely be a flurry
			//   of warnings. So, the feature would quickly need an enhancement
			//   to de-dup such warnings. (read: added complexity)
			// see also https://github.com/vmware-tanzu/carvel-kbld/issues/160
		}
	}
	return nil
}
