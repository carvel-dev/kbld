// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"sigs.k8s.io/yaml"
)

type Origin struct {
	Git         *OriginGit         `json:"git,omitempty"`
	Local       *OriginLocal       `json:"local,omitempty"`
	Resolved    *OriginResolved    `json:"resolved,omitempty"`
	Tagged      *OriginTagged      `json:"tagged,omitempty"`
	Preresolved *OriginPreresolved `json:"preresolved,omitempty"`
}

type OriginGit struct {
	RemoteURL string   `json:"remoteURL"`
	SHA       string   `json:"sha"`
	Dirty     bool     `json:"dirty"`
	Tags      []string `json:"tags,omitempty"`
}

type OriginLocal struct {
	Path string `json:"path"`
}

type OriginResolved struct {
	URL string `json:"url"`
	Tag string `json:"tag,omitempty"`
}

type OriginTagged struct {
	Tags []string `json:"tags"`
}

type OriginPreresolved struct {
	URL string `json:"url"`
}

func NewOriginsFromString(str string) ([]Origin, error) {
	var origins []Origin

	// Ignores unknown types of origin. At this time...
	// - "Origin" are provided as primarily optional diagnostic information
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
	err := yaml.Unmarshal([]byte(str), &origins)
	if err != nil {
		return []Origin{}, err
	}

	return origins, nil
}
