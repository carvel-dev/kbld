package maven

import (
	"bytes"
	ctlbdk "carvel.dev/kbld/pkg/kbld/builder/docker"
	"carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
)

var defaultImageTag = "latest"
var ImageID = regexp.MustCompile("(sha256:)([0-9a-z]+)")

type Jib struct {
	docker ctlbdk.Docker
	logger ctllog.Logger
}

func NewMavenJib(docker ctlbdk.Docker, logger ctllog.Logger) Jib {
	return Jib{docker: docker, logger: logger}
}

func (b *Jib) Run(image, directory string, opts config.SourceJibRunOpts) (ctlbdk.TmpRef, error) {

	prefixedLogger := b.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using kbld jib build): %s\n", directory)))
	defer prefixedLogger.Write([]byte("finished build (using kbld jib build)\n"))

	tag := opts.Tag
	if tag == nil {
		tag = &defaultImageTag
	}
	targetImage := fmt.Sprintf("%s:%s", image, *tag)

	var stdoutBuf, stderrBuf bytes.Buffer

	if opts.Target == nil {
		return ctlbdk.TmpRef{}, fmt.Errorf("Expected target to be specified, but was not")
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

	prefixedLogger.Write([]byte(fmt.Sprintf("running command: %s\n", cmd)))

	if err := cmd.Run(); err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	inspectData, err := b.docker.Inspect(targetImage)
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	prefixedLogger.Write([]byte(fmt.Sprintf("digest: %s, id: %s\n", inspectData.RepoDigests, inspectData.ID)))

	return b.docker.RetagStable(ctlbdk.NewTmpRef(inspectData.ID), image, inspectData.ID, prefixedLogger)
}
