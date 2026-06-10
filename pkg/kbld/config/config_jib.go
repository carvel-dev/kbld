// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

// SourceJibOpts represents configuration options for Jib sources.
type SourceJibOpts struct {
	Run SourceJibRunOpts
}

// SourceJibRunOpts represents options for running Jib.
type SourceJibRunOpts struct {
	Target     *string   `json:"target"`
	RawOptions *[]string `json:"rawOptions"`
	Tag        *string   `json:"tag,omitempty"`
}
