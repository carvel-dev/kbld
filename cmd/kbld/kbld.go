// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
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

	options := cmd.NewKbldOptions(confUI)
	command := cmd.NewKbldCmd(options)

	err := command.Execute()
	if err != nil {
		multiLineErr := uierrs.NewMultiLineError(err)
		if options.UIFlags.JSON {
			_, _ = fmt.Fprintf(os.Stderr, "kbld: Error: %s\n", multiLineErr)
		} else {
			confUI.ErrorLinef("kbld: Error: %s", multiLineErr)
		}
		os.Exit(1)
	}

	confUI.PrintLinef("Succeeded")
}
