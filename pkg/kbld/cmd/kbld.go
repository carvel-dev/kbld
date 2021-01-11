// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"io"

	"github.com/cppforlife/cobrautil"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kbld/pkg/kbld/version"
	"github.com/spf13/cobra"
)

type KbldOptions struct {
	ui      *ui.ConfUI
	UIFlags UIFlags
}

func NewKbldOptions(ui *ui.ConfUI) *KbldOptions {
	return &KbldOptions{ui: ui}
}

func NewDefaultKbldCmd(ui *ui.ConfUI) *cobra.Command {
	return NewKbldCmd(NewKbldOptions(ui))
}

func NewKbldCmd(o *KbldOptions) *cobra.Command {
	cmd := NewResolveCmd(NewResolveOptions(o.ui))

	cmd.Use = "kbld"
	cmd.Short = "kbld prepares Docker images to deploy to Kubernetes"
	cmd.Version = version.Version

	// Affects children as well
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Disable docs header
	cmd.DisableAutoGenTag = true

	// TODO bash completion
	cmd.SetOutput(uiBlockWriter{o.ui}) // setting output for cmd.Help()

	o.UIFlags.Set(cmd)

	cmd.AddCommand(NewInspectCmd(NewInspectOptions(o.ui)))
	cmd.AddCommand(NewPackageCmd(NewPackageOptions(o.ui)))
	cmd.AddCommand(NewUnpackageCmd(NewUnpackageOptions(o.ui)))
	cmd.AddCommand(NewVersionCmd(NewVersionOptions(o.ui)))
	cmd.AddCommand(NewRelocateCmd(NewRelocateOptions(o.ui)))

	// Last one runs first
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureCmdWithSubcmd)
	cobrautil.VisitCommands(cmd, cobrautil.ReconfigureLeafCmd)

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(func(*cobra.Command, []string) error {
		o.UIFlags.ConfigureUI(o.ui)
		return nil
	}))

	cobrautil.VisitCommands(cmd, cobrautil.WrapRunEForCmd(cobrautil.ResolveFlagsForCmd))

	return cmd
}

type uiBlockWriter struct {
	ui ui.UI
}

var _ io.Writer = uiBlockWriter{}

func (w uiBlockWriter) Write(p []byte) (n int, err error) {
	w.ui.PrintBlock(p)
	return len(p), nil
}
