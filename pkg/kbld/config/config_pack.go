// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

type SourcePackOpts struct {
	Build SourcePackBuildOpts
}

type SourcePackBuildOpts struct {
	Builder    *string
	Buildpacks *[]string
	ClearCache *bool     `json:"clearCache"`
	RawOptions *[]string `json:"rawOptions"`
}
