// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceBazelOpts struct {
	Build SourceBazelBuildOpts
}

type SourceBazelBuildOpts struct {
	Label      *string   `json:"label"`
	RawOptions *[]string `json:"rawOptions"`
}
