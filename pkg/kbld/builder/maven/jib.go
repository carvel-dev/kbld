// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

// Package maven implements a builder using Maven/Jib.
package maven

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	ctlbdk "carvel.dev/kbld/pkg/kbld/builder/docker"
	"carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
)

var defaultImageTag = "latest"

// Jib represents a Maven Jib builder.
type Jib struct {
	docker ctlbdk.Docker
	logger ctllog.Logger
}

// NewMavenJib creates a new Jib builder.
func NewMavenJib(docker ctlbdk.Docker, logger ctllog.Logger) Jib {
	return Jib{docker: docker, logger: logger}
}

// Run executes the Maven Jib build.
func (b *Jib) Run(image, directory string,
	opts config.SourceJibRunOpts) (ctlbdk.TmpRef, error) {
	prefixedLogger := b.logger.NewPrefixedWriter(image + " | ")

	_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
		"starting build (using kbld jib build): %s\n", directory)))
	defer func() {
		_, _ = prefixedLogger.Write([]byte(
			"finished build (using kbld jib build)\n"))
	}()

	tag := defaultImageTag
	if opts.Tag != nil {
		tag = *opts.Tag
	}
	targetImage := fmt.Sprintf("%s:%s", image, tag)

	var stdoutBuf, stderrBuf bytes.Buffer

	if opts.Target == nil {
		return ctlbdk.TmpRef{},
			errors.New("Expected target to be specified, but was not")
	}

	// Base arguments for the Maven Jib command.
	cmdArgs := []string{
		"compile",
		"jib:dockerBuild",
		"-Dimage=" + targetImage,
		"-Djib.allowInsecureRegistries=true",
	}

	if opts.RawOptions != nil {
		cmdArgs = append(cmdArgs, *opts.RawOptions...)
	}

	cmd := exec.Command("mvn", cmdArgs...)

	cmd.Dir = filepath.Join(directory, *opts.Target)
	cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
	cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

	_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
		"running command: %s\n", cmd)))

	if err := cmd.Run(); err != nil {
		_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
			"error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	inspectData, err := b.docker.Inspect(targetImage)
	if err != nil {
		_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
			"inspect error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
		"digest: %s, id: %s\n", inspectData.RepoDigests, inspectData.ID)))

	return b.docker.RetagStable(
		ctlbdk.NewTmpRef(inspectData.ID), image, inspectData.ID, prefixedLogger)
}
