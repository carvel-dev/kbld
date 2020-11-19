// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	"github.com/k14s/kbld/pkg/kbld/imagedesc"
	ctllog "github.com/k14s/kbld/pkg/kbld/logger"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
	"github.com/spf13/cobra"
)

type PackageOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
	OutputPath    string
	Concurrency   int
}

var _ imagedesc.Registry = ctlreg.Registry{}

func NewPackageOptions(ui ui.UI) *PackageOptions {
	return &PackageOptions{ui: ui}
}

func NewPackageCmd(o *PackageOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "package",
		Aliases: []string{"pkg"},
		Short:   "Package configuration and images into tarball",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", "", "Output tarball path")
	cmd.Flags().IntVar(&o.Concurrency, "concurrency", 5, "Set maximum number of concurrent imports")
	return cmd
}

func (o *PackageOptions) Run() error {
	if len(o.OutputPath) == 0 {
		return fmt.Errorf("Expected 'output' flag to be non-empty")
	}

	logger := ctllog.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("package | ")

	rs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := FindImages(rs, conf)
	if err != nil {
		return err
	}

	registry, err := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	if err != nil {
		return err
	}

	imageSet := TarImageSet{ImageSet{o.Concurrency, prefixedLogger}, o.Concurrency, prefixedLogger}

	return imageSet.Export(foundImages, o.OutputPath, registry)
}

func FindImages(allRs []ctlres.Resource, conf ctlconf.Conf) (*UnprocessedImageURLs, error) {

	foundImages := NewUnprocessedImageURLs()

	for _, res := range allRs {
		imageRefs := ctlser.NewImageRefs(res.DeepCopyRaw(), conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			foundImages.Add(UnprocessedImageURL{imgURL})
			return "", false
		})
	}

	// Include preresolved images since we want to
	// be able to package up lock files.
	for _, override := range conf.ImageOverrides() {
		if override.Preresolved {
			foundImages.Add(UnprocessedImageURL{override.NewImage})
		}
	}

	for _, img := range foundImages.All() {
		_, err := regname.NewDigest(img.URL)
		if err != nil {
			return nil, fmt.Errorf("Expected image '%s' to be in digest form (i.e. image@digest)", img.URL)
		}
	}

	return foundImages, nil
}
