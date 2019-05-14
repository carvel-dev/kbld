package cmd

import (
	"io"

	"github.com/cppforlife/cobrautil"
	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	"github.com/spf13/cobra"
)

type KbldOptions struct {
	ui            *ui.ConfUI
	configFactory cmdcore.ConfigFactory
	depsFactory   cmdcore.DepsFactory

	UIFlags         UIFlags
	KubeconfigFlags cmdcore.KubeconfigFlags
}

func NewKbldOptions(ui *ui.ConfUI, configFactory cmdcore.ConfigFactory, depsFactory cmdcore.DepsFactory) *KbldOptions {
	return &KbldOptions{ui: ui, configFactory: configFactory, depsFactory: depsFactory}
}

func NewDefaultKbldCmd(ui *ui.ConfUI) *cobra.Command {
	configFactory := cmdcore.NewConfigFactoryImpl()
	depsFactory := cmdcore.NewDepsFactoryImpl(configFactory)
	options := NewKbldOptions(ui, configFactory, depsFactory)
	flagsFactory := cmdcore.NewFlagsFactory(configFactory, depsFactory)
	return NewKbldCmd(options, flagsFactory)
}

func NewKbldCmd(o *KbldOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := NewResolveCmd(NewResolveOptions(o.ui, o.depsFactory), flagsFactory)

	cmd.Use = "kbld"
	cmd.Short = "kbld prepares Docker images to deploy to Kubernetes"

	// Affects children as well
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	// Disable docs header
	cmd.DisableAutoGenTag = true

	// TODO bash completion
	cmd.SetOutput(uiBlockWriter{o.ui}) // setting output for cmd.Help()

	o.UIFlags.Set(cmd, flagsFactory)
	o.KubeconfigFlags.Set(cmd, flagsFactory)

	o.configFactory.ConfigurePathResolver(o.KubeconfigFlags.Path.Value)
	o.configFactory.ConfigureContextResolver(o.KubeconfigFlags.Context.Value)

	cmd.AddCommand(NewInspectCmd(NewInspectOptions(o.ui, o.depsFactory), flagsFactory))
	cmd.AddCommand(NewPackageCmd(NewPackageOptions(o.ui, o.depsFactory), flagsFactory))
	cmd.AddCommand(NewUnpackageCmd(NewUnpackageOptions(o.ui, o.depsFactory), flagsFactory))
	cmd.AddCommand(NewWebsiteCmd(NewWebsiteOptions()))
	cmd.AddCommand(NewVersionCmd(NewVersionOptions(o.ui), flagsFactory))

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
