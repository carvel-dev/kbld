// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	ctlbdk "github.com/k14s/kbld/pkg/kbld/builder/docker"
	"github.com/k14s/kbld/pkg/kbld/config"
	ctllog "github.com/k14s/kbld/pkg/kbld/logger"
)

var (
	bazelImageID = regexp.MustCompile("Loaded image ID: (sha256:)?([0-9a-z]+)")
)

type Bazel struct {
	docker ctlbdk.Docker
	logger ctllog.Logger
}

func NewBazel(docker ctlbdk.Docker, logger ctllog.Logger) Bazel {
	return Bazel{docker: docker, logger: logger}
}

func (b *Bazel) Build(image, directory string, opts config.SourceBazelBuildOpts) (ctlbdk.DockerTmpRef, error) {

	prefixedLogger := b.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using bazel): %s\n", directory)))
	defer prefixedLogger.Write([]byte("finished build (using bazel)\n"))

	var imageID string
	{
		var stdoutBuf, stderrBuf bytes.Buffer

		if opts.Label == nil {
			return ctlbdk.DockerTmpRef{}, fmt.Errorf("Expected label to be specified, but was not")
		}

		cmdArgs := []string{"run"}

		if opts.Label != nil {
			cmdArgs = append(cmdArgs, *opts.Label)
		}
		if opts.RawOptions != nil {
			cmdArgs = append(cmdArgs, *opts.RawOptions...)
		}

		cmd := exec.Command("bazel", cmdArgs...)
		cmd.Dir = directory
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return ctlbdk.DockerTmpRef{}, err
		}

		matches := bazelImageID.FindStringSubmatch(stdoutBuf.String())
		if len(matches) != 3 {
			return ctlbdk.DockerTmpRef{}, fmt.Errorf("Expected to find image ID in bazel output but did not")
		}

		imageID = "sha256:" + matches[2]
	}

	return b.docker.RetagStable(ctlbdk.NewDockerTmpRef(imageID), image, imageID, prefixedLogger)
}
