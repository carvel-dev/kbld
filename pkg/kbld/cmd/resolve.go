// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"carvel.dev/imgpkg/pkg/imgpkg/lockconfig"
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlimg "carvel.dev/kbld/pkg/kbld/image"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
	ctlreg "carvel.dev/kbld/pkg/kbld/registry"
	ctlres "carvel.dev/kbld/pkg/kbld/resources"
	ctlser "carvel.dev/kbld/pkg/kbld/search"
	"carvel.dev/kbld/pkg/kbld/version"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type ResolveOptions struct {
	ui ui.UI

	FileFlags         FileFlags
	RegistryFlags     RegistryFlags
	AllowedToBuild    bool
	BuildConcurrency  int
	ImagesAnnotation  bool
	OriginsAnnotation bool
	ImageMapFile      string
	LockOutput        string
	ImgpkgLockOutput  string
	UnresolvedInspect bool
	Platform          string
}

func NewResolveOptions(ui ui.UI) *ResolveOptions {
	return &ResolveOptions{ui: ui}
}

func NewResolveCmd(o *ResolveOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Build images and update references",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().BoolVar(&o.AllowedToBuild, "build", true, "Allow building of images")
	cmd.Flags().IntVar(&o.BuildConcurrency, "build-concurrency", 4, "Set maximum number of concurrent builds")
	cmd.Flags().BoolVar(&o.ImagesAnnotation, "images-annotation", true, "Annotate resources with images annotation")
	cmd.Flags().BoolVar(&o.OriginsAnnotation, "origins-annotation", true, "Include origins annotation")
	cmd.Flags().StringVar(&o.ImageMapFile, "image-map-file", "", "Set image map file (/cnab/app/relocation-mapping.json in CNAB)")
	cmd.Flags().StringVar(&o.LockOutput, "lock-output", "", "File path to emit configuration with resolved image references")
	cmd.Flags().StringVar(&o.ImgpkgLockOutput, "imgpkg-lock-output", "", "File path to emit images lockfile with resolved image references")
	cmd.Flags().BoolVar(&o.UnresolvedInspect, "unresolved-inspect", false, "List image references found in inputs")
	cmd.Flags().StringVar(&o.Platform, "platform", "", "Apply platform selection to image indexes")
	return cmd
}

func (o *ResolveOptions) Run() error {
	if o.ImgpkgLockOutput != "" && o.LockOutput != "" {
		return fmt.Errorf("Can only output one lockfile type, please provide only one of '--lock-output' or '--imgpkg-lock-output'")
	}
	logger := ctllog.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("resolve | ")

	resBss, err := o.ResolveResources(&logger, prefixedLogger)
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

func (o *ResolveOptions) ResolveResources(logger *ctllog.Logger, pLogger *ctllog.PrefixWriter) ([][]byte, error) {
	nonConfigRs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return nil, err
	}

	conf, err = o.withImageMapConf(conf)
	if err != nil {
		return nil, err
	}

	registry, err := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	if err != nil {
		return nil, err
	}

	opts := ctlimg.FactoryOpts{
		Conf:           conf,
		AllowedToBuild: o.AllowedToBuild,
	}
	if len(o.Platform) > 0 {
		opts.GlobalPlatformSelection, err = NewPlatformSelection(o.Platform)
		if err != nil {
			return nil, err
		}
	}
	imgFactory := ctlimg.NewFactory(opts, registry, *logger)

	imageURLs, err := o.collectImageReferences(nonConfigRs, conf)
	if err != nil {
		return nil, err
	}

	if o.UnresolvedInspect {
		output, err := imageURLs.Bytes()
		if err != nil {
			return nil, err
		}
		o.ui.PrintBlock(output)
		return nil, nil
	}

	resolvedImages, err := o.resolveImages(imageURLs, imgFactory)
	if err != nil {
		return nil, err
	}

	// Record final image transformation
	for _, pair := range resolvedImages.All() {
		pLogger.WriteStr("final: %s -> %s\n", pair.UnprocessedImageURL.URL, pair.Image.URL)
	}

	err = o.emitLockOutput(conf, resolvedImages)
	if err != nil {
		return nil, err
	}

	resBss, err := o.updateRefsInResources(nonConfigRs, conf, resolvedImages, imgFactory)
	if err != nil {
		return nil, fmt.Errorf("Updating resource references: %s", err)
	}

	return resBss, nil
}

func (o *ResolveOptions) collectImageReferences(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf) (*UnprocessedImageURLs, error) {
	imageURLs := NewUnprocessedImageURLs()

	for _, res := range nonConfigRs {
		imageRefs := ctlser.NewImageRefs(res.DeepCopyRaw(), conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			imageURLs.Add(UnprocessedImageURL{imgURL})
			return "", false
		})
	}

	return imageURLs, nil
}

func (o *ResolveOptions) resolveImages(imageURLs *UnprocessedImageURLs, imgFactory ctlimg.Factory) (*ProcessedImages, error) {
	queue := NewImageQueue(imgFactory)

	resolvedImages, err := queue.Run(imageURLs, o.BuildConcurrency)
	if err != nil {
		return nil, err
	}

	return resolvedImages, nil
}

func (o *ResolveOptions) updateRefsInResources(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf, resolvedImages *ProcessedImages,
	_ ctlimg.Factory) ([][]byte, error) {

	var errs []error
	var resBss [][]byte

	for _, res := range nonConfigRs {
		resContents := res.DeepCopyRaw()
		images := []Image{}
		imageRefs := ctlser.NewImageRefs(resContents, conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			img, found := resolvedImages.FindByURL(UnprocessedImageURL{imgURL})
			if !found {
				errs = append(errs, fmt.Errorf("Expected to find image for '%s'", imgURL))
				return "", false
			}

			if o.ImagesAnnotation {
				images = append(images, img)
			}

			return img.URL, true
		})

		resBs, err := NewResourceWithImages(resContents, images).Bytes()
		if err != nil {
			return nil, err
		}

		resBss = append(resBss, resBs)
	}

	err := errFromErrs(errs)
	if err != nil {
		return nil, err
	}

	return resBss, nil
}

func errFromErrs(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var errStrs []string
	for _, err := range errs {
		errStrs = append(errStrs, err.Error())
	}
	return fmt.Errorf("\n- %s", strings.Join(errStrs, "\n- "))
}

func (o *ResolveOptions) withImageMapConf(conf ctlconf.Conf) (ctlconf.Conf, error) {
	if len(o.ImageMapFile) == 0 {
		return conf, nil
	}

	bs, err := os.ReadFile(o.ImageMapFile)
	if err != nil {
		return ctlconf.Conf{}, err
	}

	var mapping map[string]string

	err = json.Unmarshal(bs, &mapping)
	if err != nil {
		return ctlconf.Conf{}, err
	}

	additionalConfig := ctlconf.Config{}

	for k, v := range mapping {
		additionalConfig.Overrides = append(additionalConfig.Overrides, ctlconf.ImageOverride{
			ImageRef: ctlconf.ImageRef{
				Image: k,
			},
			NewImage:    v,
			Preresolved: true,
		})
	}

	return conf.WithAdditionalConfig(additionalConfig), nil
}

func (o *ResolveOptions) emitLockOutput(conf ctlconf.Conf, resolvedImages *ProcessedImages) error {
	switch {
	case o.LockOutput != "":
		c := ctlconf.NewConfig()
		c.MinimumRequiredVersion = version.Version
		c.SearchRules = conf.SearchRulesWithoutDefaults()

		for _, urlImagePair := range resolvedImages.All() {
			c.Overrides = append(c.Overrides, ctlconf.ImageOverride{
				ImageRef: ctlconf.ImageRef{
					Image: urlImagePair.UnprocessedImageURL.URL,
				},
				NewImage:    urlImagePair.Image.URL,
				Preresolved: true,
			})
		}

		return c.WriteToFile(o.LockOutput)
	case o.ImgpkgLockOutput != "":
		iLock := lockconfig.ImagesLock{
			LockVersion: lockconfig.LockVersion{
				APIVersion: lockconfig.ImagesLockAPIVersion,
				Kind:       lockconfig.ImagesLockKind,
			},
		}
		for _, urlImagePair := range resolvedImages.All() {
			iLock.Images = append(iLock.Images, lockconfig.ImageRef{
				Image:       urlImagePair.Image.URL,
				Annotations: o.imgpkgLockAnnotations(urlImagePair),
			})
		}
		return iLock.WriteToPath(o.ImgpkgLockOutput)
	default:
		return nil
	}
}

func (o *ResolveOptions) imgpkgLockAnnotations(i ProcessedImageItem) map[string]string {
	anns := map[string]string{
		ctlconf.ImagesLockKbldID: i.UnprocessedImageURL.URL,
	}
	if o.OriginsAnnotation && len(i.Origins) > 0 {
		bs, err := yaml.Marshal(i.Origins)
		if err != nil {
			return anns
		}
		anns[ctlconf.ImagesLockKbldOrigins] = string(bs)
	}

	return anns
}

// NewPlatformSelection parses platform string into PlatformSelection
// Examples: linux/386, linux/arm/v7, linux/arm/v6
// Taken from https://github.com/google/go-containerregistry/blob/570ba6c88a5041afebd4599981d849af96f5dba9/cmd/crane/cmd/util.go#L45
func NewPlatformSelection(platform string) (*ctlconf.PlatformSelection, error) {
	result := &ctlconf.PlatformSelection{}

	parts := strings.SplitN(platform, ":", 2)
	if len(parts) == 2 {
		result.OSVersion = parts[1]
	}

	parts = strings.Split(parts[0], "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Parsing platform '%s': expected format os/arch[/variant]", platform)
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("Parsing platform '%s': too many slashes", platform)
	}

	result.OS = parts[0]
	result.Architecture = parts[1]
	if len(parts) > 2 {
		result.Variant = parts[2]
	}
	return result, nil
}
