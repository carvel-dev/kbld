package config

import "strings"

// Options for builds using Containerfiles
//
// see https://github.com/containers/common/blob/main/docs/Containerfile.5.md
type ContainerFileOpts struct {
	// Always pull images
	Pull bool
	// File containing instructions
	// Docker will use "Dockerfile" as default
	// Buildah can detect "Containerfile" or "Dockerfile" as default
	File *string
	// Option "--build-arg=K=V"
	BuildArgs map[string]string `json:"buildArgs"`
	//
	Target *string
	//
	Platforms []string
}

type SourceBuildahOpts struct {
	ContainerFileOpts
	// More options
	RawOptions *[]string `json:"rawOptions"`
}

func (opts SourceBuildahOpts) Args() []string {
	args := []string{}

	if opts.Pull {
		args = append(args, "--pull")
	}
	for arg, value := range opts.BuildArgs {
		args = append(args, "--build-arg="+arg+"="+value)
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
} // SourceBuildahOpts.Args
