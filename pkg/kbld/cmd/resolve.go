package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type ResolveOptions struct {
	ui ui.UI

	FileFlags        FileFlags
	RegistryFlags    RegistryFlags
	BuildConcurrency int
	ImagesAnnotation bool
	ImageMapFile     string
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
	cmd.Flags().IntVar(&o.BuildConcurrency, "build-concurrency", 4, "Set maximum number of concurrent builds")
	cmd.Flags().BoolVar(&o.ImagesAnnotation, "images-annotation", true, "Annotate resources with images annotation")
	cmd.Flags().StringVar(&o.ImageMapFile, "image-map-file", "", "Set image map file (/cnab/app/relocation-mapping.json in CNAB)")
	return cmd
}

func (o *ResolveOptions) Run() error {
	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("resolve | ")

	nonConfigRs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	conf, err = o.withImageMapConf(conf)
	if err != nil {
		return err
	}

	registry := ctlimg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	imgFactory := ctlimg.NewFactory(conf, registry, logger)

	resolvedImages, err := o.resolveImages(nonConfigRs, conf, imgFactory)
	if err != nil {
		return err
	}

	// Record final image transformation
	for imgURL, img := range resolvedImages {
		prefixedLogger.WriteStr("final: %s -> %s\n", imgURL, img.URL)
	}

	resBss, err := o.updateRefsInResources(nonConfigRs, conf, resolvedImages, imgFactory)
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

func (o *ResolveOptions) resolveImages(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf, imgFactory ctlimg.Factory) (map[string]Image, error) {

	foundImages := map[string]struct{}{}

	for _, res := range nonConfigRs {
		imageKVs := ImageKVs{res.DeepCopyRaw(), conf.ImageKeys()}

		imageKVs.Visit(func(val interface{}) (interface{}, bool) {
			if imgURL, ok := val.(string); ok {
				foundImages[imgURL] = struct{}{}
			}
			return nil, false
		})
	}

	queue := NewImageQueue(imgFactory)

	resolvedImages, err := queue.Run(foundImages, o.BuildConcurrency)
	if err != nil {
		return nil, err
	}

	return resolvedImages, nil
}

func (o *ResolveOptions) updateRefsInResources(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf, resolvedImages map[string]Image,
	imgFactory ctlimg.Factory) ([][]byte, error) {

	var errs []error
	var resBss [][]byte

	for _, res := range nonConfigRs {
		resContents := res.DeepCopyRaw()
		images := []Image{}
		imageKVs := ImageKVs{resContents, conf.ImageKeys()}

		imageKVs.Visit(func(val interface{}) (interface{}, bool) {
			imgURL, ok := val.(string)
			if !ok {
				return nil, false
			}

			img, found := resolvedImages[imgURL]
			if !found {
				errs = append(errs, fmt.Errorf("Expected to find image for '%s'", imgURL))
				return nil, false
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

	bs, err := ioutil.ReadFile(o.ImageMapFile)
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
