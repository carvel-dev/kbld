// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceKubectlBuildkitOpts struct {
	Build SourceKubectlBuildkitBuildOpts
}

type SourceKubectlBuildkitBuildOpts struct {
	Target     *string
	Platform   *string
	Pull       *bool
	NoCache    *bool `json:"noCache"`
	File       *string
	RawOptions *[]string `json:"rawOptions"`
}
