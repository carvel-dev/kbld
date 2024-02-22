// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"carvel.dev/kbld/pkg/kbld/cmd"
	uierrs "github.com/cppforlife/go-cli-ui/errors"
	"github.com/cppforlife/go-cli-ui/ui"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetOutput(io.Discard)

	// TODO logs
	// TODO log flags used

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	command := cmd.NewDefaultKbldCmd(confUI)

	err := command.Execute()
	if err != nil {
		confUI.ErrorLinef("kbld: Error: %s", uierrs.NewMultiLineError(err))
		os.Exit(1)
	}

	confUI.PrintLinef("Succeeded")
}
