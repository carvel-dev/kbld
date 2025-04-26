package buildah

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
)

type Buildah struct {
	logger ctllog.Logger
}

func New(logger ctllog.Logger) Buildah {
	return Buildah{logger}
}

func ensureDirectory(directory string) error {
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

func Launch(directory string, command string, cmdArgs []string, prefixedLogger *ctllog.PrefixWriter) error {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command("buildah", cmdArgs...)
	cmd.Dir = directory
	cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
	cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

	err := cmd.Run()
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return err
	}
	return nil
}

func (b Buildah) BuildAndPushImage(image, directory string, imgDst ctlconf.ImageDestination, opts ctlconf.SourceBuildahOpts) (string, error) {

	err := ensureDirectory(directory)
	if err != nil {
		return "", err
	}

	tagRef := imgDst.NewImage

	prefixedLogger := b.logger.NewPrefixedWriter(image + " build | ")
	prefixedLogger.Write([]byte(fmt.Sprintf("Start building using buildah\n")))

	cmdArgs := []string{"build", "--tag", tagRef}

	if opts.Pull {
		cmdArgs = append(cmdArgs, "--pull")
	}
	if opts.File != nil {
		cmdArgs = append(cmdArgs, "--file="+*opts.File)
	}
	if opts.Target != nil {
		cmdArgs = append(cmdArgs, "--target="+*opts.Target)
	}
	if opts.RawOptions != nil {
		cmdArgs = append(cmdArgs, *opts.RawOptions...)
	}
	// Use current directory as context
	// cmdArgs = append(cmdArgs, "./")

	build_err := Launch(directory, "buildah", cmdArgs, prefixedLogger)
	if build_err != nil {
		return "", build_err
	}
	push_err := b.PushImage(image, tagRef)
	if push_err != nil {
		return "", nil
	}
	return tagRef, nil
}

func (b Buildah) PushImage(image, tagRef string) error {
	prefixedLogger := b.logger.NewPrefixedWriter(image + " push | ")
	push_err := Launch("", "buildah", []string{"push", tagRef}, prefixedLogger)
	if push_err != nil {
		return push_err
	}
	return nil
}
