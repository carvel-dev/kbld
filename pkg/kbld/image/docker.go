package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"time"

	regname "github.com/google/go-containerregistry/pkg/name"
)

var (
	tmpRefHint = regexp.MustCompile("[^a-zA-Z0-9\\-]+")
)

type Docker struct {
	logger Logger
}

type DockerTmpRef struct {
	val string
}

func (r DockerTmpRef) AsString() string { return r.val }

type DockerImageDigest struct {
	val string
}

func (r DockerImageDigest) AsString() string { return r.val }

func (d Docker) Build(image, directory string) (DockerTmpRef, error) {
	// Timestamp at the beginning for easier sorting
	tmpRef := DockerTmpRef{fmt.Sprintf("kbld:%d-%s", time.Now().UTC().UnixNano(), tmpRefHint.ReplaceAllString(image, "-"))}
	prefixedLogger := d.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using Docker): %s -> %s\n", directory, tmpRef.AsString())))
	defer prefixedLogger.Write([]byte("finished build (using Docker)\n"))

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "build", "-t", tmpRef.AsString(), ".")
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

	// Retag image with its sha256 to produce exact image ref if nothing has changed
	// Seems that Docker doesn't like `kbld@sha256:...` format for local images
	// Image hint at the beginning for easier sorting
	stableTmpRef := DockerTmpRef{fmt.Sprintf("kbld:%s-%s",
		tmpRefHint.ReplaceAllString(image, "-"),
		tmpRefHint.ReplaceAllString(inspectData.Id, "-"))}

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

	// Remove temporary tag to be nice to `docker images` output
	{
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

func (d Docker) Push(tmpRef DockerTmpRef, imageDst string) (DockerImageDigest, error) {
	prefixedLogger := d.logger.NewPrefixedWriter(imageDst + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting push (using Docker): %s -> %s\n", tmpRef.AsString(), imageDst)))
	defer prefixedLogger.Write([]byte("finished push (using Docker)\n"))

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "tag", tmpRef.AsString(), imageDst)
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("tag error: %s\n", err)))
			return DockerImageDigest{}, err
		}
	}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "push", imageDst)
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("push error: %s\n", err)))
			return DockerImageDigest{}, err
		}
	}

	inspectData, err := d.inspect(imageDst)
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return DockerImageDigest{}, err
	}

	if len(inspectData.RepoDigests) != 1 {
		prefixedLogger.Write([]byte("repo digests mismatch\n"))
		return DockerImageDigest{}, fmt.Errorf("Expected to find exactly one repo digest, but found %#v", inspectData.RepoDigests)
	}

	ref := inspectData.RepoDigests[0]

	nameWithDigest, err := regname.NewDigest(ref, regname.WeakValidation)
	if err != nil {
		return DockerImageDigest{}, fmt.Errorf("Extracting reference digest from '%s': %s", ref, err)
	}

	return DockerImageDigest{nameWithDigest.DigestStr()}, nil
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
