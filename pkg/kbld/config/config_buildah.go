package config

// Options for builds using Containerfiles
//
// see https://github.com/containers/common/blob/main/docs/Containerfile.5.md
type ContainerFileOpts struct {
	// Always pull images
	Pull bool
	// File containing instructions
	// Docker will use "Dockerfile" as default
	// Buildah can detect "Containerfile" or "Dockerfile" as default
	File   *string
	Target *string
}

type SourceBuildahOpts struct {
	ContainerFileOpts
}
