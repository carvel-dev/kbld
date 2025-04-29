package buildah

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	ctlb "carvel.dev/kbld/pkg/kbld/builder"
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

// Push the buildah manifest and return the digest
func BuildahPush(src string, dest string, log *ctllog.PrefixWriter) (string, error) {
	digest_file, digest_err := os.CreateTemp("", "buildah-")
	if digest_err != nil {
		return "", fmt.Errorf("cannot create digest file: %w", digest_err)
	}
	defer func() {
		if err := digest_file.Close(); err != nil {
			fmt.Printf("ERROR: Closing temp file %q: %v", digest_file.Name(), err)
		}
		if err := os.Remove(digest_file.Name()); err != nil {
			fmt.Printf("ERROR: Removing temp file %q: %v", digest_file.Name(), err)
		}
	}()

	// !!! with --digestfile, buildah will not return an error if an authentication is required.
	pushCommand := exec.Command("buildah", "manifest", "push", "--all", "--digestfile="+digest_file.Name(), src, "docker://"+dest)
	pushCommand.Stdout = log
	push_err := pushCommand.Run()
	if push_err != nil {
		return "", fmt.Errorf("error pushing to %q (check if you are authenticated) : %w", dest, push_err)
	}

	digest := make([]byte, 64+7)
	digest_len, read_err := digest_file.Read(digest)
	if read_err != nil {
		return "", fmt.Errorf("cannot read digest in file %q (check if you are authenticated) : %w", digest_file.Name(), read_err)
	}
	return string(digest[0:digest_len]), nil
} //// BuildahPush

// Push built image to a remote registry
// Return the image URL with digest
func (b Buildah) PushImage(image string, imgDst *ctlconf.ImageDestination) (string, error) {
	prefixedLogger := b.logger.NewPrefixedWriter(image + " push | ")
	var remoteImg string
	if imgDst == nil {
		remoteImg = image
	} else {
		tb := ctlb.TagBuilder{}
		randSuffix, err := tb.RandomStr50()
		if err != nil {
			return "", fmt.Errorf("generating image dst suffix: %s", err)
		}
		remoteImg = imgDst.NewImage + ":kbld-" + randSuffix
	}

	digest, push_err := BuildahPush(image, remoteImg, prefixedLogger)
	if push_err != nil {
		return "", push_err
	}
	return remoteImg + "@" + digest, nil
} //// PushImage
