// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	uierrs "github.com/cppforlife/go-cli-ui/errors"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kbld/pkg/kbld/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetOutput(ioutil.Discard)

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
