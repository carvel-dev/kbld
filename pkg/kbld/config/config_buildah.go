// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"slices"
	"strings"
)

// ContainerFileOpts Options for builds using Containerfiles
//
// see https://github.com/containers/common/blob/main/docs/Containerfile.5.md
type ContainerFileOpts struct {
	// Pull Always pull images
	Pull *bool
	// NoCache Do not use cache when building the image
	NoCache *bool
	// File containing instructions
	// Docker will use "Dockerfile" as default
	// Buildah can detect "Containerfile" or "Dockerfile" as default
	File *string
	// BuildArgs Option "--build-arg=K=V"
	BuildArgs map[string]string `json:"buildArgs"`
	// Target Set the target build stage to build.
	Target *string
	// Platforms to build for
	Platforms []string
}

// SourceBuildahOpts specifies options for building images using Buildah.
type SourceBuildahOpts struct {
	ContainerFileOpts
	// More options
	RawOptions *[]string `json:"rawOptions"`
}

// Args stringifies the options
//
//revive:disable-next-line:cognitive-complexity
func (opts SourceBuildahOpts) Args() []string {
	args := []string{}

	if opts.Pull != nil && *opts.Pull {
		args = append(args, "--pull")
	}
	if opts.NoCache != nil && *opts.NoCache {
		args = append(args, "--no-cache")
	}

	if opts.BuildArgs != nil {
		args = opts.buildArgs()
	}

	if opts.Target != nil {
		args = append(args, "--target="+*opts.Target)
	}
	if len(opts.Platforms) > 0 {
		args = append(args, "--platform="+strings.Join(opts.Platforms, ","))
	}
	if opts.RawOptions != nil {
		args = append(args, *opts.RawOptions...)
	}
	return args
}

func (opts SourceBuildahOpts) buildArgs() []string {
	var args []string
	keys := make([]string, 0, len(opts.BuildArgs))
	for k := range opts.BuildArgs {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, arg := range keys {
		value := opts.BuildArgs[arg]
		args = append(args, "--build-arg="+arg+"="+value)
	}
	return args
}
