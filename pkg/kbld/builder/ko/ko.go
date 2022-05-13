// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package ko

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	ctlbdk "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/builder/docker"
	"github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctllog "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/logger"
)

type Ko struct {
	logger ctllog.Logger
}

func NewKo(logger ctllog.Logger) Ko {
	return Ko{logger: logger}
}

func (k *Ko) Build(image, directory string, opts config.SourceKoBuildOpts) (ctlbdk.TmpRef, error) {
	prefixedLogger := k.logger.NewPrefixedWriter(image + " | ")

	prefixedLogger.Write([]byte(fmt.Sprintf("starting build (using ko): %s\n", directory)))
	defer prefixedLogger.Write([]byte("finished build (using ko)\n"))

	var stdoutBuf, stderrBuf bytes.Buffer

	cmdArgs := []string{"publish", ".", "--local"}

	if opts.RawOptions != nil {
		cmdArgs = append(cmdArgs, *opts.RawOptions...)
	}

	cmd := exec.Command("ko", cmdArgs...)
	cmd.Dir = directory
	cmd.Stdout = io.MultiWriter(&stdoutBuf, prefixedLogger)
	cmd.Stderr = io.MultiWriter(&stderrBuf, prefixedLogger)

	err := cmd.Run()
	if err != nil {
		prefixedLogger.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return ctlbdk.TmpRef{}, err
	}

	return ctlbdk.NewTmpRef(strings.Trim(stdoutBuf.String(), "\n")), nil
}
