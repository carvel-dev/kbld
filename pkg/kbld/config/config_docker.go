// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceDockerOpts struct {
	Build SourceDockerBuildOpts
}

type SourceDockerBuildOpts struct {
	Target     *string
	Pull       *bool
	NoCache    *bool `json:"noCache"`
	File       *string
	Buildkit   *bool
	RawOptions *[]string `json:"rawOptions"`
}
