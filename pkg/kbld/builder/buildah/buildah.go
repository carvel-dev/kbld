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

func (b Buildah) BuildAndPushImage(image string, directory string, imgDst *ctlconf.ImageDestination, opts ctlconf.SourceBuildahOpts) (string, error) {

	err := ensureDirectory(directory)
	if err != nil {
		return "", err
	}

	prefixedLogger := b.logger.NewPrefixedWriter(image + " build | ")
	prefixedLogger.Write([]byte(fmt.Sprintf("Start building using buildah\n")))

	cmdArgs := []string{"build", "--manifest=" + image}

	if opts.Pull {
		cmdArgs = append(cmdArgs, "--pull")
	}
	if opts.File != nil {
		cmdArgs = append(cmdArgs, "--file="+*opts.File)
	}
	for arg, value := range opts.BuildArgs {
		cmdArgs = append(cmdArgs, "--build-arg="+arg+"="+value)
	}
	if opts.Target != nil {
		cmdArgs = append(cmdArgs, "--target="+*opts.Target)
	}
	if len(opts.Platforms) > 0 {
		cmdArgs = append(cmdArgs, "--platform="+strings.Join(opts.Platforms, ","))
	}

	if opts.RawOptions != nil {
		cmdArgs = append(cmdArgs, *opts.RawOptions...)
	}
	// Use current directory as context
	// cmdArgs = append(cmdArgs, "./")

	{
		cmd := exec.Command("buildah", cmdArgs...)
		cmd.Dir = directory
		cmd.Stdout = prefixedLogger

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return "", err
		}
	}
	remoteRef, push_err := b.PushImage(image, imgDst)
	if push_err != nil {
		return "", push_err
	}
	prefixedLogger.WriteStr("Image build : " + remoteRef)
	return remoteRef, nil
}

func BuildahPush(src string, dest string, log *ctllog.PrefixWriter) error {
	pushCommand := exec.Command("buildah", "manifest", "push", "--all", src, "docker://"+dest)
	pushCommand.Stdout = log
	push_err := pushCommand.Run()
	if push_err != nil {
		return push_err
	}
	return nil
} //// BuildahPush

// Push built image to a remote registry
// Return one of the remote image address
func (b Buildah) PushImage(image string, imgDst *ctlconf.ImageDestination) (string, error) {
	prefixedLogger := b.logger.NewPrefixedWriter(image + " push | ")
	if imgDst == nil {
		push_err := BuildahPush(image, image, prefixedLogger)
		if push_err != nil {
			return "", push_err
		}
		return image, nil
	} else if len(imgDst.Tags) > 0 {
		for _, tag := range imgDst.Tags {
			push_err := BuildahPush(image, imgDst.NewImage+":"+tag, prefixedLogger)
			if push_err != nil {
				return "", push_err
			}
		}
		return imgDst.NewImage + ":" + imgDst.Tags[0], nil
	} else {
		push_err := BuildahPush(image, imgDst.NewImage+":kbld", prefixedLogger)
		if push_err != nil {
			return "", push_err
		}
		return imgDst.NewImage + ":kbld", nil
	}
} //// PushImage
