// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type Meta interface {
	meta()
}

type ImageMeta struct {
	Metas []Meta
}

type BuiltImageSourceGit struct {
	Type      string // always set to 'git'
	RemoteURL string `json:",omitempty" yaml:",omitempty"`
	SHA       string
	Dirty     bool
	Tags      []string `json:",omitempty" yaml:",omitempty"`
}

type BuiltImageSourceLocal struct {
	Type string // always set to 'local'
	Path string
}

type ResolvedImageSourceURL struct {
	Type string // always set to 'resolved'
	URL  string
	Tag  string
}

type TaggedImageMeta struct {
	Type string // always set to 'tagged'
	Tags []string
}

type PreresolvedImageSourceURL struct {
	Type string // always set to 'preresolved'
	URL  string
	Tag  string `json:",omitempty" yaml:",omitempty"`
}

func (BuiltImageSourceGit) meta()       {}
func (BuiltImageSourceLocal) meta()     {}
func (ResolvedImageSourceURL) meta()    {}
func (TaggedImageMeta) meta()           {}
func (PreresolvedImageSourceURL) meta() {}

func metasHistory(metas string) ([]Meta, error) {
	imageMeta := ImageMeta{}
	err := yaml.Unmarshal([]byte(metas), &imageMeta)
	if err != nil {
		return []Meta{}, err
	}
	return imageMeta.Metas, nil
}

var _ json.Unmarshaler = &ImageMeta{}

func (m *ImageMeta) UnmarshalJSON(data []byte) error {
	var list []interface{}
	err := yaml.Unmarshal(data, &list)
	if err != nil {
		return err
	}

	var local BuiltImageSourceLocal
	var git BuiltImageSourceGit
	var res ResolvedImageSourceURL
	var preres PreresolvedImageSourceURL
	var tag TaggedImageMeta

	for _, item := range list {
		yamlItem, _ := yaml.Marshal(&item)
		switch {
		case yaml.Unmarshal(yamlItem, &local) == nil && local.Type == "local":
			m.Metas = append(m.Metas, local)
		case yaml.Unmarshal(yamlItem, &git) == nil && git.Type == "git":
			m.Metas = append(m.Metas, git)
		case yaml.Unmarshal(yamlItem, &res) == nil && res.Type == "resolved":
			m.Metas = append(m.Metas, res)
		case yaml.Unmarshal(yamlItem, &preres) == nil && preres.Type == "preresolved":
			m.Metas = append(m.Metas, preres)
		case yaml.Unmarshal(yamlItem, &tag) == nil && tag.Type == "tagged":
			m.Metas = append(m.Metas, tag)
		default:
			return fmt.Errorf("Unknown Image Meta")
		}
	}
	return nil
}
