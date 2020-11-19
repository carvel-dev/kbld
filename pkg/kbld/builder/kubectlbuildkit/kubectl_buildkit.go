// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlb "github.com/k14s/kbld/pkg/kbld/builder"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctllog "github.com/k14s/kbld/pkg/kbld/logger"
)

var (
	// Example output that includes final digest:
	//   ...
	// 	 #10 exporting layers done
	// 	 #10 exporting manifest sha256:55d863c4231ec285b88516942fd5d636216c36b6a686a20bf28d1aa5125c16b7 0.0s done
	// 	 #10 exporting config sha256:aa1fce99c57c864cc8d98f7ff4a54bc3b9b7e3f63a741de926a3393bc76f5986 0.0s done
	//   ...
	//   #10 exporting layers done
	//   #10 exporting manifest sha256:55d863c4231ec285b88516942fd5d636216c36b6a686a20bf28d1aa5125c16b7 done
	//   #10 exporting config sha256:aa1fce99c57c864cc8d98f7ff4a54bc3b9b7e3f63a741de926a3393bc76f5986 done
	//   #10 pushing layers
	kubectlBuildkitDigest = regexp.MustCompile("exporting manifest (sha256:)?([0-9a-z]+) ")
)

type KubectlBuildkit struct {
	logger ctllog.Logger
}

func NewKubectlBuildkit(logger ctllog.Logger) KubectlBuildkit {
	return KubectlBuildkit{logger}
}

func (d KubectlBuildkit) BuildAndPush(image, directory string,
	imgDst *ctlconf.ImageDestination, opts ctlconf.SourceKubectlBuildkitOpts) (string, error) {

	tagRef, err := d.tagRef(image, imgDst)
	if err != nil {
		return "", err
	}

	prefixedLogger := d.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using kubectl buildkit): %s -> %s\n", directory, tagRef)))
	defer prefixedLogger.Write([]byte("finished build (using kubectl buildkit)\n"))

	var stdoutBuf, stderrBuf bytes.Buffer

	cmdArgs := []string{"buildkit", "build", "--progress=plain"}

	if opts.Build.Target != nil {
		cmdArgs = append(cmdArgs, "--target", *opts.Build.Target)
	}
	if opts.Build.Platform != nil {
		cmdArgs = append(cmdArgs, "--platform", *opts.Build.Platform)
	}
	if opts.Build.Pull != nil && *opts.Build.Pull {
		cmdArgs = append(cmdArgs, "--pull")
	}
	if opts.Build.NoCache != nil && *opts.Build.NoCache {
		cmdArgs = append(cmdArgs, "--no-cache")
	}
	if opts.Build.File != nil {
		// Since docker command is executed with cwd of directory,
		// Dockerfile path doesnt need to be joined with it
		cmdArgs = append(cmdArgs, "--file", *opts.Build.File)
	}
	if opts.Build.RawOptions != nil {
		cmdArgs = append(cmdArgs, *opts.Build.RawOptions...)
	}

	if imgDst != nil {
		cmdArgs = append(cmdArgs, "--push")
		// https://github.com/vmware-tanzu/buildkit-cli-for-kubectl/blob/main/docs/multiarch.md#using-a-registry
		// > it's possible to skip specifying the --registry-secret flag to kubectl build by naming the secret the same name as the builder
	}

	cmdArgs = append(cmdArgs, "--tag", tagRef, ".")

	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Dir = directory
	cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
	cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

	err = cmd.Run()
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return "", err
	}

	// Exercise digest finding logic regardless of pushing or not
	digestMatches := kubectlBuildkitDigest.FindStringSubmatch(stderrBuf.String())
	if len(digestMatches) != 3 {
		return "", fmt.Errorf("Expected to find image digest in build output but did not")
	}

	// Since kubectl builtkit project targets both containerd and Docker daemon
	// and Docker daemon does not support use of digests for locally loaded images
	// we can only return digest ref when image is pushed to registry
	if imgDst != nil {
		digestRefStr := imgDst.NewImage + "@sha256:" + digestMatches[2]

		digestRef, err := regname.NewDigest(digestRefStr, regname.WeakValidation)
		if err != nil {
			return "", fmt.Errorf("Validating destination digest ref '%s': %s", digestRefStr, err)
		}

		return digestRef.Name(), nil
	}

	return tagRef, nil
}

func (d KubectlBuildkit) tagRef(image string, imgDst *ctlconf.ImageDestination) (string, error) {
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
