// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceBazelOpts struct {
	Run SourceBazelRunOpts
}

type SourceBazelRunOpts struct {
	Target     *string   `json:"target"`
	RawOptions *[]string `json:"rawOptions"`
}
