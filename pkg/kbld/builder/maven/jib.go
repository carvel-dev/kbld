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

	ctlb "carvel.dev/kbld/pkg/kbld/builder"
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

	tmpRef, err := b.tmpRef(image)
	if err != nil {
		return ctlbdk.TmpRef{}, err
	}

	if opts.Target == nil {
		return ctlbdk.TmpRef{},
			errors.New("Expected target to be specified, but was not")
	}

	err = b.runMaven(directory, tmpRef.AsString(), opts, prefixedLogger)
	if err != nil {
		_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
			"error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	inspectData, err := b.docker.Inspect(tmpRef.AsString())
	if err != nil {
		_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
			"inspect error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
		"digest: %s, id: %s\n", inspectData.RepoDigests, inspectData.ID)))

	stableTmpRef, err := b.docker.RetagStable(
		tmpRef, image, inspectData.ID, prefixedLogger)
	if err != nil {
		return ctlbdk.TmpRef{}, err
	}

	err = b.tagTarget(stableTmpRef, targetImage, prefixedLogger)
	if err != nil {
		_, _ = prefixedLogger.Write([]byte(fmt.Sprintf(
			"target tag error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	return stableTmpRef, nil
}

func (*Jib) tmpRef(image string) (ctlbdk.TmpRef, error) {
	tb := ctlb.TagBuilder{}
	randPrefix50, err := tb.RandomStr50()
	if err != nil {
		return ctlbdk.TmpRef{},
			fmt.Errorf("Generating tmp image suffix: %s", err)
	}

	return ctlbdk.NewTmpRef("kbld:" + tb.CheckTagLen128(fmt.Sprintf(
		"%s-%s",
		randPrefix50,
		tb.TrimStr(tb.CleanStr(image), 50),
	))), nil
}

func (*Jib) runMaven(directory string, targetImage string,
	opts config.SourceJibRunOpts, prefixedLogger *ctllog.PrefixWriter) error {
	var stdoutBuf, stderrBuf bytes.Buffer

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
		"running command: %v\n", cmd)))

	return cmd.Run()
}

func (*Jib) tagTarget(stableTmpRef ctlbdk.TmpRef, targetImage string,
	prefixedLogger *ctllog.PrefixWriter) error {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("docker", "tag",
		stableTmpRef.AsString(), targetImage)
	cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
	cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

	return cmd.Run()
}
