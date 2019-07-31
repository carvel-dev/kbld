package image

import (
	"bytes"
	"crypto/rand"
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
	randPrefix, err := d.randomStr(5)
	if err != nil {
		return DockerTmpRef{}, fmt.Errorf("Generating tmp image suffix: %s", err)
	}

	tmpRef := DockerTmpRef{fmt.Sprintf(
		"kbld:%s-%s", randPrefix, tmpRefHint.ReplaceAllString(image, "-"))}
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

	// Retag image with its sha256 to produce exact image ref if nothing has changed.
	// Seems that Docker doesn't like `kbld@sha256:...` format for local images.
	// Image hint at the beginning for easier sorting.
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

	// Remove temporary tag to be nice to `docker images` output.
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

	// Generate random tag for pushed image.
	// TODO we are technically polluting registry with new tags.
	// Unfortunately we do not know digest upfront so cannot use kbld-sha256-... format.
	imageDstTagged, err := regname.NewTag(imageDst, regname.WeakValidation)
	if err == nil {
		randSuffix, err := d.randomStr(5)
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Generating image dst suffix: %s", err)
		}

		imageDstTag := fmt.Sprintf("kbld-%s", randSuffix)

		imageDstTagged, err = regname.NewTag(imageDst+":"+imageDstTag, regname.WeakValidation)
		if err != nil {
			return DockerImageDigest{}, fmt.Errorf("Generating image dst tag '%s': %s", imageDst, err)
		}
	}

	imageDst = imageDstTagged.Name()

	prefixedLogger.Write([]byte(fmt.Sprintf("starting push (using Docker): %s -> %s\n", tmpRef.AsString(), imageDst)))
	defer prefixedLogger.Write([]byte("finished push (using Docker)\n"))

	prevInspectData, err := d.inspect(tmpRef.AsString())
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return DockerImageDigest{}, err
	}

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

	currInspectData, err := d.inspect(imageDst)
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

func (d Docker) randomStr(n int) (string, error) {
	bs, err := d.randomBytes(n)
	if err != nil {
		return "", err
	}
	result := ""
	for _, b := range bs {
		result += fmt.Sprintf("%d", b)
	}
	// Timestamp at the beginning for easier sorting
	return fmt.Sprintf("rand-%d-%s", time.Now().UTC().UnixNano(), result), nil
}

func (d Docker) randomBytes(n int) ([]byte, error) {
	bs := make([]byte, n)
	_, err := rand.Read(bs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
