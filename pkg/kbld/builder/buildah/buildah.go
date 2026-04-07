// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

// Package buildah enable
package buildah

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	ctlb "carvel.dev/kbld/pkg/kbld/builder"
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
)

// Buildah struct to define the Builder
type Buildah struct {
	logger ctllog.Logger
}

// New create a new Buildah builder
func New(logger ctllog.Logger) Buildah {
	return Buildah{logger}
}

func ensureDirectory(directory string) error {
	stat, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf(
			"Checking if path '%s' is a directory: %s", directory, err)
	}

	// Provide explicit directory check error message because otherwise
	// docker CLI outputs confusing msg
	//   'error: fork/exec /usr/local/bin/docker: not a directory'
	if !stat.IsDir() {
		return fmt.Errorf(
			"Expected path '%s' to be a directory, but was not", directory)
	}

	return nil
}

// Generate a name to store the image in local
// The local name is always new and random.
// The manifest is new each time and do not accumulate images.
func localImageName(configImageName string,
	imgDest *ctlconf.ImageDestination) string {
	if imgDest != nil {
		configImageName = imgDest.NewImage
	}
	tb := ctlb.TagBuilder{}
	randSuffix, err := tb.RandomStr50()
	if err != nil {
		return configImageName + ":kbld"
	}
	return configImageName + ":kbld-" + randSuffix
}

// BuildAndPushImage builds and pushed the images to the registry
func (b Buildah) BuildAndPushImage(image string, directory string,
	imgDst *ctlconf.ImageDestination,
	opts ctlconf.SourceBuildahOpts) (string, error) {
	if imgDst == nil {
		return "", errors.New(
			"a destination is required to store the built image")
	}

	err := ensureDirectory(directory)
	if err != nil {
		return "", err
	}

	prefixedLogger := b.logger.NewPrefixedWriter(image + " build | ")
	prefixedLogger.Write([]byte("Start building using buildah\n"))

	localName := localImageName(image, imgDst)
	cmdArgs := []string{"build", "--manifest=" + localName}

	if opts.File != nil {
		cmdArgs = append(cmdArgs, "--file="+*opts.File)
	}
	cmdArgs = append(cmdArgs, opts.Args()...)

	// Use current directory as context
	// cmdArgs = append(cmdArgs, "./")

	prefixedLogger.WriteStr("=> buildah " + strings.Join(cmdArgs, " "))
	{
		cmd := exec.Command("buildah", cmdArgs...)
		cmd.Dir = directory
		cmd.Stdout = prefixedLogger
		cmd.Stderr = io.MultiWriter(os.Stderr, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return "", err
		}
	}

	pushLogger := b.logger.NewPrefixedWriter(image + " push | ")
	// Push using a temporary, random tag,
	// and return a canonical digest reference.
	tempTagBytes := make([]byte, 8)
	if _, err := rand.Read(tempTagBytes); err != nil {
		return "", fmt.Errorf("generating temporary tag: %w", err)
	}
	tempTag := hex.EncodeToString(tempTagBytes)
	tempRemoteName := fmt.Sprintf("%s:%s", imgDst.NewImage, tempTag)
	digest, pushErr := Push(localName, tempRemoteName, pushLogger)

	if pushErr != nil {
		return "", pushErr
	}
	remoteName := fmt.Sprintf("%s@%s", imgDst.NewImage, digest)
	prefixedLogger.WriteStr("Image build : " + remoteName)
	return remoteName, nil
}

// Push the buildah manifest and return the digest
func Push(src string, dest string,
	log *ctllog.PrefixWriter) (string, error) {
	digestFile, digestErr := os.CreateTemp("", "buildah-")
	if digestErr != nil {
		return "", fmt.Errorf(
			"cannot create digest file: %w", digestErr)
	}
	defer func() {
		if err := digestFile.Close(); err != nil {
			//revive:disable-next-line:unhandled-error
			fmt.Printf(
				"ERROR: Closing temp file %q: %v", digestFile.Name(), err)
		}
		if err := os.Remove(digestFile.Name()); err != nil {
			//revive:disable-next-line:unhandled-error
			fmt.Printf(
				"ERROR: Removing temp file %q: %v", digestFile.Name(), err)
		}
	}()

	// !!! with --digestfile, buildah will not return an error
	// even if an authentication is required.
	log.WriteStr("=> buildah manifest push --all --digestfile=" +
		digestFile.Name() + " " + src + " docker://" + dest)
	pushCommand := exec.Command("buildah", "manifest", "push", "--all",
		"--digestfile="+digestFile.Name(), src, "docker://"+dest)
	pushCommand.Stdout = log
	pushCommand.Stderr = log
	pushErr := pushCommand.Run()
	if pushErr != nil {
		return "", fmt.Errorf(
			"error pushing to %q (check if you are authenticated) : %w",
			dest, pushErr)
	}

	digest, err := calculateDigest(digestFile)
	if err != nil {
		return "", err
	}
	return digest, nil
}

func calculateDigest(digestFile *os.File) (string, error) {
	digestBytes, readErr := os.ReadFile(digestFile.Name())
	if readErr != nil {
		//revive:disable-next-line:line-length-limit
		return "", fmt.Errorf("cannot read digest in file %q (check if you are authenticated) : %w",
			digestFile.Name(), readErr)
	}

	digest := strings.TrimSpace(string(digestBytes))
	if digest == "" {
		return "", fmt.Errorf(
			"no digest found in file %q (check if you are authenticated)",
			digestFile.Name())
	}
	digestPattern := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	if !digestPattern.MatchString(digest) {
		return "", fmt.Errorf(
			"invalid digest format %q in file %q", digest, digestFile.Name())
	}
	return digest, nil
}
