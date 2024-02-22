// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	ctlb "carvel.dev/kbld/pkg/kbld/builder"
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
	regname "github.com/google/go-containerregistry/pkg/name"
)

/*

Useful docs:
- https://cloudolife.com/2022/03/05/Infrastructure-as-Code-IaC/Container/Docker/Docker-buildx-support-multiple-architectures-images/

Commands:
- docker buildx create --name mybuilder
- docker buildx use mybuilder
- docker buildx inspect --bootstrap
or
- docker buildx create --name mybuilder --use

Example:

$ docker buildx build . -t k14stest/test-image:buildx-test-img --load --progress=plain
...snip...
#10 exporting to oci image format
#10 exporting layers done
#10 exporting manifest sha256:cd6d662fcf06854810c83e4bff197ef52e65b39c13baa143107ebd32cc6f8636 0.0s done
#10 exporting config sha256:e3bdd21522d99c37f355c70bef342c99cb1e19ee4cf8014001d750a73a5b8d42 done
#10 sending tarball
#10 sending tarball 0.1s done
#10 DONE 0.2s
#11 importing to docker
#11 DONE 0.0s

$ docker buildx build . -t k14stest/test-image:buildx-test-img --push --progress=plain
...snip...
#10 [auth] k14stest/test-image:pull,push token for registry-1.docker.io
#10 DONE 0.0s
#11 exporting to image
#11 exporting layers done
#11 exporting manifest sha256:cd6d662fcf06854810c83e4bff197ef52e65b39c13baa143107ebd32cc6f8636 done
#11 exporting config sha256:e3bdd21522d99c37f355c70bef342c99cb1e19ee4cf8014001d750a73a5b8d42 done
#11 pushing layers
#11 pushing layers 0.3s done
#11 pushing manifest for docker.io/k14stest/test-image:buildx-test-img@sha256:cd6d662fcf06854810c83e4bff197ef52e65b39c13baa143107ebd32cc6f8636
#11 pushing manifest for docker.io/k14stest/test-image:buildx-test-img@sha256:cd6d662fcf06854810c83e4bff197ef52e65b39c13baa143107ebd32cc6f8636 0.2s done
#11 DONE 0.6s

$ docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 . --push -t k14stest/test-image:buildx-test-img --progress=plain
...snip...
#22 exporting manifest sha256:32d9f4784456eb53e77768d94a705d1432b833368496d15280311b5879d54ef6 done
#22 exporting config sha256:f9d9b90fceeb8c721ced16025b812613c7f52e5b507e2b2ae39b3d424381725c done
#22 exporting manifest sha256:a18f53367186f5af99bf3d629a45522638e95feb5807b926713d6004232d0df2 done
#22 exporting config sha256:29d5c3d35c3e9262111de2575c97d63f8bda9ddab69f88a2143f1fb77a259f4d 0.0s done
#22 exporting manifest list sha256:8db4e06667f35cc30b101936f2f8f4a6a846f71ef4a7e4112602fd324f63fd65 done
#22 pushing layers
#22 pushing layers 0.5s done
#22 pushing manifest for docker.io/k14stest/test-image:buildx-test-img@sha256:8db4e06667f35cc30b101936f2f8f4a6a846f71ef4a7e4112602fd324f63fd65
#22 pushing manifest for docker.io/k14stest/test-image:buildx-test-img@sha256:8db4e06667f35cc30b101936f2f8f4a6a846f71ef4a7e4112602fd324f63fd65 0.4s done
#22 DONE 0.9s

*/

var (
	// not worried too much with hard coding sha256 for forseeable future
	dockerBuildxPushDigest = regexp.MustCompile("pushing manifest for .*@sha256:([0-9a-z]+) ")
	dockerBuildxPushErr    = "does not currently support exporting manifest lists"
)

type Buildx struct {
	docker Docker
	logger ctllog.Logger
}

func NewBuildx(docker Docker, logger ctllog.Logger) Buildx {
	return Buildx{docker, logger}
}

// BuildAndOptionallyPush either loads built image into Docker daemon
// or pushes it to specified registry.
func (d Buildx) BuildAndOptionallyPush(
	image, directory string, imgDst *ctlconf.ImageDestination,
	opts ctlconf.SourceDockerBuildxOpts) (string, error) {

	err := d.ensureDirectory(directory)
	if err != nil {
		return "", err
	}

	tagRef, err := d.tagRef(image, imgDst)
	if err != nil {
		return "", err
	}

	prefixedLogger := d.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using Docker buildx): %s -> %s\n", directory, tagRef)))
	defer prefixedLogger.Write([]byte("finished build (using Docker buildx)\n"))

	var stdoutBuf, stderrBuf bytes.Buffer
	{
		cmdArgs := []string{"buildx", "build", "--progress=plain"}

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

		cmdArgs = append(cmdArgs, "--tag", tagRef, ".")

		// Load built image into Docker daemon, otherwise it's not being used anywhere
		if imgDst != nil {
			cmdArgs = append(cmdArgs, "--push")
		} else {
			cmdArgs = append(cmdArgs, "--load")
		}

		cmd := exec.Command("docker", cmdArgs...)
		cmd.Dir = directory
		cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
		cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

		err := cmd.Run()
		if err != nil {
			if strings.Contains(stderrBuf.String(), dockerBuildxPushErr) {
				prefixedLogger.Write([]byte("(hint: Specify image destination as multi-platform builds are not supported on local Docker)\n"))
			}
			prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return "", err
		}
	}

	if imgDst != nil {
		// Digest is only printed when push option was selected
		digestMatches := dockerBuildxPushDigest.FindStringSubmatch(stderrBuf.String())
		if len(digestMatches) != 2 {
			return "", fmt.Errorf("Expected to find image digest in build output but did not")
		}

		digestRefStr := imgDst.NewImage + "@sha256:" + digestMatches[1]

		digestRef, err := regname.NewDigest(digestRefStr, regname.WeakValidation)
		if err != nil {
			return "", fmt.Errorf("Validating destination digest ref '%s': %s", digestRefStr, err)
		}

		return digestRef.Name(), nil
	}

	// Work with locally stored image in Docker daemon
	inspectData, err := d.docker.Inspect(tagRef)
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("inspect error: %s\n", err)))
		return "", err
	}

	tmpRef, err := d.docker.RetagStable(TmpRef{tagRef}, image, inspectData.ID, prefixedLogger)
	if err != nil {
		return "", err
	}

	return tmpRef.AsString(), nil
}

func (d Buildx) tagRef(image string, imgDst *ctlconf.ImageDestination) (string, error) {
	tb := ctlb.TagBuilder{}

	randPrefix50, err := tb.RandomStr50()
	if err != nil {
		return "", fmt.Errorf("Generating tmp image suffix: %s", err)
	}

	tag := tb.CheckTagLen128(fmt.Sprintf(
		"%s-%s",
		randPrefix50,
		tb.TrimStr(tb.CleanStr(image), 50),
	))

	if imgDst != nil {
		tagRef := imgDst.NewImage + ":" + tag

		_, err := regname.NewTag(tagRef, regname.WeakValidation)
		if err != nil {
			return "", fmt.Errorf("Validating destination tag ref '%s': %s", tagRef, err)
		}

		return tagRef, nil
	}

	return "kbld:" + tag, nil
}

func (d Buildx) ensureDirectory(directory string) error {
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
