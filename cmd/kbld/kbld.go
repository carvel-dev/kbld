package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kbld/pkg/kbld/cmd"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
		confUI.ErrorLinef("Error: %v", err)
		os.Exit(1)
	}

	confUI.PrintLinef("Succeeded")
}
