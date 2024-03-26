// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceDockerOpts struct {
	Build  SourceDockerBuildOpts
	Buildx *SourceDockerBuildxOpts
}

type SourceDockerBuildOpts struct {
	Target     *string
	Pull       *bool
	NoCache    *bool `json:"noCache"`
	File       *string
	Buildkit   *bool
	RawOptions *[]string `json:"rawOptions"`
}

type SourceDockerBuildxOpts struct {
	Target     *string
	Pull       *bool
	NoCache    *bool `json:"noCache"`
	File       *string
	RawOptions *[]string `json:"rawOptions"`
}
