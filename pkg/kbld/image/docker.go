// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	regname "github.com/google/go-containerregistry/pkg/name"
)

var (
	tmpRefHint = regexp.MustCompile("[^a-zA-Z0-9\\-]+")
)

type Docker struct {
	logger Logger
}

type DockerBuildOpts struct {
	// https://docs.docker.com/engine/reference/commandline/build/
	Target     *string
	Pull       *bool
	NoCache    *bool
	File       *string
	RawOptions *[]string
}

type DockerTmpRef struct {
	val string
}

func (r DockerTmpRef) AsString() string { return r.val }

type DockerImageDigest struct {
	val string
}

func (r DockerImageDigest) AsString() string { return r.val }

func (d Docker) Build(image, directory string, opts DockerBuildOpts) (DockerTmpRef, error) {
	err := d.ensureDirectory(directory)
	if err != nil {
		return DockerTmpRef{}, err
	}

	randPrefix50, err := d.randomStr50()
	if err != nil {
		return DockerTmpRef{}, fmt.Errorf("Generating tmp image suffix: %s", err)
	}

	tmpRef := DockerTmpRef{d.checkTagLen128(fmt.Sprintf(
		"kbld:%s-%s",
		randPrefix50,
		d.trimStr(tmpRefHint.ReplaceAllString(image, "-"), 50),
	))}

	prefixedLogger := d.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using Docker): %s -> %s\n", directory, tmpRef.AsString())))
	defer prefixedLogger.Write([]byte("finished build (using Docker)\n"))

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmdArgs := []string{"build"}

		if opts.Target != nil {
			cmdArgs = append(cmdArgs, "--target", *opts.Target)
		}
		if opts.Pull != nil && *opts.Pull {
			cmdArgs = append(cmdArgs, "--pull")
		}
		if opts.NoCache != nil && *opts.NoCache {
			cmdArgs = append(cmdArgs, "--no-cache")
		}
		if opts.File != nil {
			// Since docker command is executed with cwd of directory,
			// Dockerfile path doesnt need to be joined with it
			cmdArgs = append(cmdArgs, "--file", *opts.File)
		}
		if opts.RawOptions != nil {
			cmdArgs = append(cmdArgs, *opts.RawOptions...)
		}

		cmdArgs = append(cmdArgs, "--tag", tmpRef.AsString(), ".")

		cmd := exec.Command("docker", cmdArgs...)
		cmd.Dir = directory
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return DockerTmpRef{}, err
		}
	}

	inspectData, err := d.inspect(tmpRef.AsString())
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return DockerTmpRef{}, err
	}

	return d.RetagStable(tmpRef, image, inspectData.Id, prefixedLogger)
}

func (d Docker) RetagStable(tmpRef DockerTmpRef, image, imageID string,
	prefixedLogger *LoggerPrefixWriter) (DockerTmpRef, error) {

	// Retag image with its sha256 to produce exact image ref if nothing has changed.
	// Seems that Docker doesn't like `kbld@sha256:...` format for local images.
	// Image hint at the beginning for easier sorting.
	stableTmpRef := DockerTmpRef{d.checkTagLen128(fmt.Sprintf(
		"kbld:%s-%s",
		d.trimStr(tmpRefHint.ReplaceAllString(image, "-"), 50),
		d.checkLen(tmpRefHint.ReplaceAllString(imageID, "-"), 72),
	))}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "tag", tmpRef.AsString(), stableTmpRef.AsString())
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("tag error: %s\n", err)))
			return DockerTmpRef{}, err
		}
	}

	// Remove temporary tag to be nice to `docker images` output.
	// (No point in "untagging" digest reference)
	if !strings.HasPrefix(tmpRef.AsString(), "sha256:") {
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "rmi", tmpRef.AsString())
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("untag error: %s\n", err)))
			return DockerTmpRef{}, err
		}
	}

	return stableTmpRef, nil
}

func (d Docker) Push(tmpRef DockerTmpRef, imageDst string, additionalImageTags []string) (DockerImageDigest, error) {
	prefixedLogger := d.logger.NewPrefixedWriter(imageDst + " | ")

	// Generate random tag for pushed image.
	// TODO we are technically polluting registry with new tags.
	// Unfortunately we do not know digest upfront so cannot use kbld-sha256-... format.
	imageDstTagged, err := regname.NewTag(imageDst, regname.WeakValidation)
	if err == nil {
		randSuffix, err := d.randomStr50()
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Generating image dst suffix: %s", err)
		}

		imageDstTag := fmt.Sprintf("kbld-%s", randSuffix)

		imageDstTagged, err = regname.NewTag(imageDst+":"+imageDstTag, regname.WeakValidation)
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Generating image dst tag '%s': %s", imageDst, err)
		}
	}

	for i, imageDstTag := range additionalImageTags {
		imageDstTag, err := regname.NewTag(imageDst+":"+imageDstTag, regname.WeakValidation)
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Generating image dst tag '%s': %s", imageDstTag, err)
		}
		additionalImageTags[i] = imageDstTag.Name()
	}

	fullImageDst := imageDstTagged.Name()

	additionalImageTags = append(additionalImageTags, fullImageDst)

	prefixedLogger.Write([]byte(fmt.Sprintf("starting push (using Docker): %s -> %s\n", tmpRef.AsString(), fullImageDst)))
	defer prefixedLogger.Write([]byte("finished push (using Docker)\n"))

	prevInspectData, err := d.inspect(tmpRef.AsString())
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return DockerImageDigest{}, err
	}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		for _, imageDstTag := range additionalImageTags {
			cmd := exec.Command("docker", "tag", tmpRef.AsString(), imageDstTag)
			cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
			cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

			err := cmd.Run()
			if err != nil {
				prefixedLogger.Write([]byte(fmt.Sprintf("tag error: %s\n", err)))
				return DockerImageDigest{}, err
			}
		}
	}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		for _, imageTag := range additionalImageTags {
			cmd := exec.Command("docker", "push", imageTag)
			cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
			cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

			err := cmd.Run()
			if err != nil {
				prefixedLogger.Write([]byte(fmt.Sprintf("push error: %s\n", err)))
				return DockerImageDigest{}, err
			}
		}
	}

	currInspectData, err := d.inspect(fullImageDst)
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return DockerImageDigest{}, err
	}

	// Try to detect if image we should be pushing isnt the one we ended up pushing
	// given that its theoretically possible concurrent Docker commands
	// may have retagged in the middle of the process.
	if prevInspectData.Id != currInspectData.Id {
		prefixedLogger.Write([]byte(fmt.Sprintf("push race error: %s\n", err)))
		return DockerImageDigest{}, err
	}

	return d.determineRepoDigest(currInspectData, prefixedLogger)
}

func (d Docker) ensureDirectory(directory string) error {
	stat, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("Checking if path '%s' is a directory: %s", directory, err)
	}

	// Provide explicit directory check error message because otherwise docker CLI
	// outputs confusing msg 'error: fork/exec /usr/local/bin/docker: not a directory'
	if !stat.IsDir() {
		return fmt.Errorf("Expected path '%s' to be a directory, but was not", directory)
	}

	return nil
}

func (d Docker) determineRepoDigest(inspectData dockerInspectData, prefixedLogger *LoggerPrefixWriter) (DockerImageDigest, error) {
	if len(inspectData.RepoDigests) == 0 {
		prefixedLogger.Write([]byte("missing repo digest\n"))
		return DockerImageDigest{}, fmt.Errorf("Expected to find at least one repo digest")
	}

	digestStrs := map[string]struct{}{}

	for _, rd := range inspectData.RepoDigests {
		nameWithDigest, err := regname.NewDigest(rd, regname.WeakValidation)
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Extracting reference digest from '%s': %s", rd, err)
		}
		digestStrs[nameWithDigest.DigestStr()] = struct{}{}
	}

	if len(digestStrs) != 1 {
		prefixedLogger.Write([]byte("repo digests mismatch\n"))
		return DockerImageDigest{}, fmt.Errorf("Expected to find same repo digest, but found %#v", inspectData.RepoDigests)
	}

	for digest, _ := range digestStrs {
		return DockerImageDigest{digest}, nil
	}

	panic("unreachable")
}

type dockerInspectData struct {
	Id          string
	RepoDigests []string
}

func (d Docker) inspect(ref string) (dockerInspectData, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("docker", "inspect", ref)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		return dockerInspectData{}, err
	}

	var data []dockerInspectData

	err = json.Unmarshal(stdoutBuf.Bytes(), &data)
	if err != nil {
		return dockerInspectData{}, err
	}

	if len(data) != 1 {
		return dockerInspectData{}, fmt.Errorf("Expected to find exactly one image, but found %d", len(data))
	}

	return data[0], nil
}

func (d Docker) checkTagLen128(tag string) string {
	// "A tag ... may contain a maximum of 128 characters."
	// (https://docs.docker.com/engine/reference/commandline/tag/)
	return d.checkLen(tag, 128)
}

func (d Docker) checkLen(str string, num int) string {
	if len(str) > num {
		panic(fmt.Sprintf("Expected string '%s' len to be less than %d", str, num))
	}
	return str
}

func (d Docker) trimStr(str string, num int) string {
	if len(str) > num {
		str = str[:num]
		// Do not end strings on dash
		if strings.HasSuffix(str, "-") {
			str = str[:len(str)-1] + "e"
		}
	}
	return str
}

func (d Docker) randomStr50() (string, error) {
	bs, err := d.randomBytes(5)
	if err != nil {
		return "", err
	}
	result := ""
	for _, b := range bs {
		result += fmt.Sprintf("%d", b)
	}
	// Timestamp at the beginning for easier sorting
	return d.checkLen(fmt.Sprintf("rand-%d-%s", time.Now().UTC().UnixNano(), result), 50), nil
}

func (d Docker) randomBytes(n int) ([]byte, error) {
	bs := make([]byte, n)
	_, err := rand.Read(bs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
