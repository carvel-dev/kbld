// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceJibOpts struct {
	Run SourceJibRunOpts
}

type SourceJibRunOpts struct {
	Target     *string   `json:"target"`
	RawOptions *[]string `json:"rawOptions"`
}
