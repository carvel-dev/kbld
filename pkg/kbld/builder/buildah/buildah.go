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

		err := cmd.Run()
		if err != nil {
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return "", err
		}
	}
	remoteRef, pushErr := b.PushImage(image, imgDst)
	if pushErr != nil {
		return "", pushErr
	}
	prefixedLogger.WriteStr("Image build : " + remoteRef)
	return remoteRef, nil
}

// Push the buildah manifest and return the digest
func BuildahPush(src string, dest string, log *ctllog.PrefixWriter) (string, error) {
	digestFile, digestErr := os.CreateTemp("", "buildah-")
	if digestErr != nil {
		return "", fmt.Errorf("cannot create digest file: %w", digestErr)
	}
	defer func() {
		if err := digestFile.Close(); err != nil {
			fmt.Printf("ERROR: Closing temp file %q: %v", digestFile.Name(), err)
		}
		if err := os.Remove(digestFile.Name()); err != nil {
			fmt.Printf("ERROR: Removing temp file %q: %v", digestFile.Name(), err)
		}
	}()

	// !!! with --digestfile, buildah will not return an error if an authentication is required.
	log.WriteStr("=> buildah manifest push --all --digestfile=" + digestFile.Name() + " " + src + " docker://" + dest)
	pushCommand := exec.Command("buildah", "manifest", "push", "--all", "--digestfile="+digestFile.Name(), src, "docker://"+dest)
	pushCommand.Stdout = log
	pushErr := pushCommand.Run()
	if pushErr != nil {
		return "", fmt.Errorf("error pushing to %q (check if you are authenticated) : %w", dest, pushErr)
	}

	digest := make([]byte, 64+7)
	digestLen, readErr := digestFile.Read(digest)
	if readErr != nil {
		return "", fmt.Errorf("cannot read digest in file %q (check if you are authenticated) : %w", digestFile.Name(), readErr)
	}
	return string(digest[0:digestLen]), nil
} // BuildahPush

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

	digest, pushErr := BuildahPush(image, remoteImg, prefixedLogger)
	if pushErr != nil {
		return "", pushErr
	}
	return remoteImg + "@" + digest, nil
} // PushImage
