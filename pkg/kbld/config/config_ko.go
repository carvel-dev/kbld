// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

type SourceKoOpts struct {
	Build SourceKoBuildOpts
}

type SourceKoBuildOpts struct {
	RawOptions *[]string `json:"rawOptions"`
}
