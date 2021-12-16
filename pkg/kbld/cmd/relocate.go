// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
	ctllog "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/logger"
	ctlreg "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/registry"
	ctlres "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/resources"
	ctlser "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/search"
	"github.com/vmware-tanzu/carvel-kbld/pkg/kbld/version"
	"sigs.k8s.io/yaml"
)

type RelocateOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
	Repository    string
	LockOutput    string
	Concurrency   int
}

func NewRelocateOptions(ui ui.UI) *RelocateOptions {
	return &RelocateOptions{ui: ui}
}

func NewRelocateCmd(o *RelocateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "relocate",
		Long: `
Command "relocate" is deprecated, please use 'imgpkg copy', learn more in https://carvel.dev/imgpkg/docs/latest/commands/#copy

Relocate images between two registries
`,
		Short:  "Relocate images between two registries",
		Hidden: true,
		RunE:   func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().StringVarP(&o.Repository, "repository", "r", "", "Import images into given image repository (e.g. docker.io/dkalinin/my-project)")
	cmd.Flags().StringVar(&o.LockOutput, "lock-output", "", "File path to emit configuration with resolved image references")
	cmd.Flags().IntVar(&o.Concurrency, "concurrency", 5, "Set maximum number of concurrent imports")
	return cmd
}

func (o *RelocateOptions) Run() error {
	logger := ctllog.NewLogger(os.Stderr)
	warningLogger := logger.NewPrefixedWriter("Warning: ")
	err := warningLogger.WriteStr(`Command "relocate" is deprecated, please use 'imgpkg copy', learn more in https://carvel.dev/imgpkg/docs/latest/commands/#copy`)
	if err != nil {
		return err
	}

	// basic checks
	if len(o.Repository) == 0 {
		return fmt.Errorf("Expected repository flag to be non-empty")
	}

	if len(o.FileFlags.Files) == 0 {
		return fmt.Errorf("Expected at least one input file")
	}

	prefixedLogger := logger.NewPrefixedWriter("relocate | ")

	// get resources from files
	rs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := FindImages(rs, conf)
	if err != nil {
		return err
	}

	importRepo, err := regname.NewRepository(o.Repository)
	if err != nil {
		return fmt.Errorf("Building import repository ref: %s", err)
	}

	dstRegistry, err := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	if err != nil {
		return err
	}

	imageSet := ImageSet{o.Concurrency, prefixedLogger}

	importedImages, err := imageSet.Relocate(foundImages, importRepo, dstRegistry)
	if err != nil {
		return err
	}

	err = o.emitLockOutput(conf, importedImages)
	if err != nil {
		return err
	}

	// Update previous image references with new references
	resBss, err := o.updateRefsInResources(rs, conf, importedImages)
	if err != nil {
		return err
	}

	// Print all resources as one YAML stream
	for _, resBs := range resBss {
		resBs = append([]byte("---\n"), resBs...)
		o.ui.PrintBlock(resBs)
	}

	return nil
}

func (o *RelocateOptions) updateRefsInResources(
	nonConfigRs []ctlres.Resource, conf ctlconf.Conf,
	resolvedImages *ProcessedImages) ([][]byte, error) {

	var missingImageErrs []error
	var resBss [][]byte

	for _, res := range nonConfigRs {
		resContents := res.DeepCopyRaw()
		imageRefs := ctlser.NewImageRefs(resContents, conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			outputImg, found := resolvedImages.FindByURL(UnprocessedImageURL{imgURL})
			if found {
				return outputImg.URL, true
			}
			missingImageErrs = append(missingImageErrs, fmt.Errorf("Expected to find image for '%s'", imgURL))
			return "", false
		})

		resBs, err := yaml.Marshal(resContents)
		if err != nil {
			return nil, err
		}

		resBss = append(resBss, resBs)
	}

	err := errFromErrs(missingImageErrs)
	if err != nil {
		return nil, err
	}

	return resBss, nil
}

func (o *RelocateOptions) emitLockOutput(conf ctlconf.Conf, resolvedImages *ProcessedImages) error {
	if len(o.LockOutput) == 0 {
		return nil
	}

	c := ctlconf.NewConfig()
	c.MinimumRequiredVersion = version.Version
	c.SearchRules = conf.SearchRulesWithoutDefaults()

	for _, override := range conf.ImageOverrides() {
		if override.Preresolved {
			img, found := resolvedImages.FindByURL(UnprocessedImageURL{override.NewImage})
			if !found {
				return fmt.Errorf("Expected to find imported image for '%s'", override.NewImage)
			}

			c.Overrides = append(c.Overrides, ctlconf.ImageOverride{
				ImageRef:    override.ImageRef,
				NewImage:    img.URL,
				Preresolved: true,
			})
		}
	}

	// TODO should we dedup overrides?
	for _, urlImagePair := range resolvedImages.All() {
		c.Overrides = append(c.Overrides, ctlconf.ImageOverride{
			ImageRef: ctlconf.ImageRef{
				Image: urlImagePair.UnprocessedImageURL.URL,
			},
			NewImage:    urlImagePair.Image.URL,
			Preresolved: true,
		})
	}

	c.Overrides = ctlconf.UniqueImageOverrides(c.Overrides)

	return c.WriteToFile(o.LockOutput)
}
