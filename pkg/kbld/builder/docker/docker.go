// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	ctlb "carvel.dev/kbld/pkg/kbld/builder"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
	regname "github.com/google/go-containerregistry/pkg/name"
)

type Docker struct {
	logger ctllog.Logger
}

type BuildOpts struct {
	// https://docs.docker.com/engine/reference/commandline/build/
	Target     *string
	Pull       *bool
	NoCache    *bool
	File       *string
	Buildkit   *bool
	RawOptions *[]string
}

type TmpRef struct {
	val string
}

func NewTmpRef(val string) TmpRef {
	return TmpRef{val}
}

func (r TmpRef) AsString() string { return r.val }

type ImageDigest struct {
	val string
}

func (r ImageDigest) AsString() string { return r.val }

func New(logger ctllog.Logger) Docker {
	return Docker{logger}
}

func (d Docker) Build(image, directory string, opts BuildOpts) (TmpRef, error) {
	err := d.ensureDirectory(directory)
	if err != nil {
		return TmpRef{}, err
	}

	tb := ctlb.TagBuilder{}

	randPrefix50, err := tb.RandomStr50()
	if err != nil {
		return TmpRef{}, fmt.Errorf("Generating tmp image suffix: %s", err)
	}

	tmpRef := TmpRef{"kbld:" + tb.CheckTagLen128(fmt.Sprintf(
		"%s-%s",
		randPrefix50,
		tb.TrimStr(tb.CleanStr(image), 50),
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

		if opts.Buildkit != nil {
			cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
		}

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return TmpRef{}, err
		}
	}

	inspectData, err := d.Inspect(tmpRef.AsString())
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return TmpRef{}, err
	}

	return d.RetagStable(tmpRef, image, inspectData.ID, prefixedLogger)
}

func (d Docker) RetagStable(tmpRef TmpRef, image, imageID string,
	prefixedLogger *ctllog.PrefixWriter) (TmpRef, error) {

	tb := ctlb.TagBuilder{}

	// Retag image with its sha256 to produce exact image ref if nothing has changed.
	// Seems that Docker doesn't like `kbld@sha256:...` format for local images.
	// Image hint at the beginning for easier sorting.
	stableTmpRef := TmpRef{"kbld:" + tb.CheckTagLen128(fmt.Sprintf(
		"%s-%s",
		tb.TrimStr(tb.CleanStr(image), 50),
		tb.CheckLen(tb.CleanStr(imageID), 72),
	))}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "tag", tmpRef.AsString(), stableTmpRef.AsString())
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("tag error: %s\n", err)))
			return TmpRef{}, err
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
			return TmpRef{}, err
		}
	}

	return stableTmpRef, nil
}

func (d Docker) Push(tmpRef TmpRef, imageDst string) (ImageDigest, error) {
	prefixedLogger := d.logger.NewPrefixedWriter(imageDst + " | ")

	tb := ctlb.TagBuilder{}

	// Generate random tag for pushed image.
	// TODO we are technically polluting registry with new tags.
	// Unfortunately we do not know digest upfront so cannot use kbld-sha256-... format.
	imageDstTagged, err := regname.NewTag(imageDst, regname.WeakValidation)
	badNameErr := (*regname.ErrBadName)(nil)
	if err == nil {
		randSuffix, err := tb.RandomStr50()
		if err != nil {
			return ImageDigest{}, fmt.Errorf("Generating image dst suffix: %s", err)
		}

		imageDstTag := fmt.Sprintf("kbld-%s", randSuffix)

		imageDstTagged, err = regname.NewTag(imageDst+":"+imageDstTag, regname.WeakValidation)
		if err != nil {
			return ImageDigest{}, fmt.Errorf(
				"Generating image dst tag '%s': %s", imageDst, err)
		}
	} else if errors.As(err, &badNameErr) {
		imageDstTagged, err = regname.NewTag(lowerCaseRepository(imageDst))
		if err != nil {
			newError := fmt.Errorf(
				"Lower casing repository '%s' still failed: %w", imageDst, err)
			return ImageDigest{}, newError
		}
	}

	imageDst = imageDstTagged.Name()

	prefixedLogger.Write([]byte(fmt.Sprintf("starting push (using Docker): %s -> %s\n", tmpRef.AsString(), imageDst)))
	defer prefixedLogger.Write([]byte("finished push (using Docker)\n"))

	prevInspectData, err := d.Inspect(tmpRef.AsString())
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return ImageDigest{}, err
	}

	{
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := exec.Command("docker", "tag", tmpRef.AsString(), imageDst)
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("tag error: %s\n", err)))
			return ImageDigest{}, err
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
			return ImageDigest{}, err
		}
	}

	currInspectData, err := d.Inspect(imageDst)
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return ImageDigest{}, err
	}

	// Try to detect if image we should be pushing isnt the one we ended up pushing
	// given that its theoretically possible concurrent Docker commands
	// may have retagged in the middle of the process.
	if prevInspectData.ID != currInspectData.ID {
		prefixedLogger.Write([]byte(fmt.Sprintf("push race error: %s\n", err)))
		return ImageDigest{}, err
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

func (d Docker) determineRepoDigest(inspectData InspectData,
	prefixedLogger *ctllog.PrefixWriter) (ImageDigest, error) {

	if len(inspectData.RepoDigests) == 0 {
		prefixedLogger.Write([]byte("missing repo digest\n"))
		return ImageDigest{}, fmt.Errorf("Expected to find at least one repo digest")
	}

	digestStrs := map[string]struct{}{}

	for _, rd := range inspectData.RepoDigests {
		nameWithDigest, err := regname.NewDigest(rd, regname.WeakValidation)
		if err != nil {
			return ImageDigest{}, fmt.Errorf("Extracting reference digest from '%s': %s", rd, err)
		}
		digestStrs[nameWithDigest.DigestStr()] = struct{}{}
	}

	if len(digestStrs) != 1 {
		prefixedLogger.Write([]byte("repo digests mismatch\n"))
		return ImageDigest{}, fmt.Errorf("Expected to find same repo digest, but found %#v", inspectData.RepoDigests)
	}

	for digest := range digestStrs {
		return ImageDigest{digest}, nil
	}

	panic("unreachable")
}

type InspectData struct {
	ID          string
	RepoDigests []string
}

func (d Docker) Inspect(ref string) (InspectData, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("docker", "inspect", ref)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	_, _ = d.logger.NewPrefixedWriter(ref + " | ").Write(
		[]byte(fmt.Sprintf("running command: %v\n", cmd)))

	err := cmd.Run()
	if err != nil {
		return InspectData{}, err
	}

	var data []InspectData

	err = json.Unmarshal(stdoutBuf.Bytes(), &data)
	if err != nil {
		return InspectData{}, err
	}

	if len(data) != 1 {
		return InspectData{}, fmt.Errorf("Expected to find exactly one image, but found %d", len(data))
	}

	return data[0], nil
}

const (
	notFound          = -1
	registrySeparator = "/"
	registryDotDelim  = "."
	registryPortDelim = ":"
)

// LowerCaseRegistry lowercases only the registry hostname portion of an
// image reference (preserving repository path and tag case), so that
// callers outside this package can normalize refs before comparing them,
// consistent with RFC 1035/1123 hostname case-insensitivity.
func LowerCaseRegistry(ref string) string {
	return lowerCaseRepository(ref)
}

// lowerCaseRepository lowercases only the registry hostname while preserving
// repository path and tag case (per RFC 1035/1123 case-insensitive names).
func lowerCaseRepository(ref string) string {
	parsedRef, err := regname.ParseReference(ref)
	if err == nil {
		return tryLowercaseRegistryFromParsed(ref, parsedRef)
	}

	lowerRef := strings.ToLower(ref)
	parsedLower, err := regname.ParseReference(lowerRef)
	if err == nil {
		return tryLowercaseRegistryFromLowercased(ref, parsedLower)
	}

	return lowercaseRegistryBySlash(ref)
}

// tryLowercaseRegistryFromParsed handles the case where the original ref
// parses successfully, checking if the registry is explicit in the original.
func tryLowercaseRegistryFromParsed(ref string,
	parsedRef regname.Reference) string {
	// Determine if registry is explicit or implicit by checking if it appears
	// literally in the original ref (case-insensitive).
	refLower := strings.ToLower(ref)
	registryStr := strings.ToLower(parsedRef.Context().RegistryStr())

	// For explicit registries, look for the registry followed by a slash.
	// This distinguishes "docker.io/lib..." (explicit docker.io) from
	// "library/image" (implicit docker.io registry, "library" is namespace).
	registryWithSlash := registryStr + registrySeparator
	registryIdx := strings.Index(refLower, registryWithSlash)

	if registryIdx == notFound {
		// Try to detect explicit registry without the canonical suffix.
		// E.g., "docker.io" appears in input but canonicalizes to
		// "index.docker.io". Look for the first "/" and check if before it
		// looks like a registry.
		firstSlash := strings.Index(refLower, registrySeparator)
		if firstSlash == notFound {
			// No slash, bare image name (implicit registry).
			return ref
		}

		// There's a slash. Check if part before it contains dots or colons
		// (registry indicators).
		beforeSlash := refLower[:firstSlash]
		hasDot := strings.Contains(beforeSlash, registryDotDelim)
		hasColon := strings.Contains(beforeSlash, registryPortDelim)
		if !hasDot && !hasColon {
			// Looks like a namespace, not a registry (implicit registry).
			return ref
		}

		// Looks like an explicit registry. Lowercase just the registry part.
		return strings.ToLower(ref[:firstSlash]) + ref[firstSlash:]
	}

	// Registry found with canonical suffix. Extract repository from original.
	repoStartIdx := registryIdx + len(registryWithSlash)
	originalRepo := ref[repoStartIdx:]

	// Reconstruct with lowercased registry and original-case repository.
	return registryStr + registrySeparator + originalRepo
}

// tryLowercaseRegistryFromLowercased handles the case where parsing the
// original ref failed but parsing the lowercased version succeeds.
func tryLowercaseRegistryFromLowercased(ref string,
	_ regname.Reference) string {
	firstSlash := strings.Index(ref, registrySeparator)
	if firstSlash == notFound {
		// No slash, bare image name (implicit registry).
		return ref
	}

	// Check if the part before "/" is a registry (has dots or colons) or
	// a namespace.
	beforeSlash := ref[:firstSlash]
	hasDot := strings.Contains(beforeSlash, registryDotDelim)
	hasColon := strings.Contains(beforeSlash, registryPortDelim)

	if !hasDot && !hasColon {
		// Looks like a namespace, not a registry (implicit registry).
		// Registry is implicit, preserve case completely.
		return ref
	}

	// Looks like an explicit registry. Lowercase just the registry part.
	return strings.ToLower(beforeSlash) + ref[firstSlash:]
}

// lowercaseRegistryBySlash is the fallback when both parsing attempts fail.
// When neither ParseReference(ref) nor ParseReference(strings.ToLower(ref))
// succeeds, use heuristic detection and aggressive case fixing to maintain
// backward compatibility without throwing errors during error recovery.
func lowercaseRegistryBySlash(ref string) string {
	firstSlash := strings.Index(ref, registrySeparator)
	if firstSlash == notFound {
		// No slash, bare image name without explicit registry.
		return ref
	}

	beforeSlash := ref[:firstSlash]

	// Check if beforeSlash looks like a hostname (registry).
	hasPort := strings.Contains(beforeSlash, registryPortDelim)
	hasDot := strings.Contains(beforeSlash, registryDotDelim)
	isLocalhost := strings.EqualFold(beforeSlash, "localhost")

	if hasPort || hasDot || isLocalhost {
		// Looks like a registry hostname, lowercase it.
		result := strings.ToLower(beforeSlash) + ref[firstSlash:]
		// If repo path still has uppercase (e.g., MyOrg/MyApp), fall back to
		// fully lowercase to ensure parsing succeeds in error recovery.
		repoHasUppercase := strings.ContainsAny(result[firstSlash:],
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		if repoHasUppercase {
			return strings.ToLower(ref)
		}
		return result
	}

	// Looks like a namespace/repository name. If it has uppercase that would
	// cause parsing to fail, lowercase everything to recover.
	if strings.ContainsAny(ref, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return strings.ToLower(ref)
	}

	return ref
}
