// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"sigs.k8s.io/yaml"
)

type Meta struct {
	Git         *MetaGit         `json:"git,omitempty"`
	Local       *MetaLocal       `json:"local,omitempty"`
	Resolved    *MetaResolved    `json:"resolved,omitempty"`
	Tagged      *MetaTagged      `json:"tagged,omitempty"`
	Preresolved *MetaPreresolved `json:"preresolved,omitempty"`
}

type MetaGit struct {
	RemoteURL string   `json:"remoteURL"`
	SHA       string   `json:"sha"`
	Dirty     bool     `json:"dirty"`
	Tags      []string `json:"tags,omitempty"`
}

type MetaLocal struct {
	Path string `json:"path"`
}

type MetaResolved struct {
	URL string `json:"url"`
	Tag string `json:"tag,omitempty"`
}

type MetaTagged struct {
	Tags []string `json:"tags"`
}

type MetaPreresolved struct {
	URL string `json:"url"`
	Tag string `json:"tag,omitempty"`
}

func NewMetasFromString(str string) ([]Meta, error) {
	var metas []Meta

	// Ignores unknown types of meta. At this time...
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
	err := yaml.Unmarshal([]byte(str), &metas)
	if err != nil {
		return []Meta{}, err
	}

	return metas, nil
}
