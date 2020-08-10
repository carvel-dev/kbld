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
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
	"github.com/k14s/kbld/pkg/kbld/version"
	"github.com/spf13/cobra"
)

type ResolveOptions struct {
	ui ui.UI

	FileFlags        FileFlags
	RegistryFlags    RegistryFlags
	BuildConcurrency int
	ImagesAnnotation bool
	ImageMapFile     string
	LockOutput       string
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
	cmd.Flags().StringVar(&o.LockOutput, "lock-output", "", "File path to emit configuration with resolved image references")
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

	registry, err := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	if err != nil {
		return err
	}

	imgFactory := ctlimg.NewFactory(conf, registry, logger)

	resolvedImages, err := o.resolveImages(nonConfigRs, conf, imgFactory)
	if err != nil {
		return err
	}

	// Record final image transformation
	for _, pair := range resolvedImages.All() {
		prefixedLogger.WriteStr("final: %s -> %s\n", pair.UnprocessedImageURL.URL, pair.Image.URL)
	}

	err = o.emitLockOutput(conf, resolvedImages)
	if err != nil {
		return err
	}

	resBss, err := o.updateRefsInResources(nonConfigRs, conf, resolvedImages, imgFactory)
	if err != nil {
		return fmt.Errorf("Updating resource references: %s", err)
	}

	// Print all resources as one YAML stream
	for _, resBs := range resBss {
		resBs = append([]byte("---\n"), resBs...)
		o.ui.PrintBlock(resBs)
	}

	return nil
}

func (o *ResolveOptions) resolveImages(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf, imgFactory ctlimg.Factory) (*ProcessedImages, error) {

	imageURLs := NewUnprocessedImageURLs()

	for _, res := range nonConfigRs {
		imageRefs := ctlser.NewImageRefs(res.DeepCopyRaw(), conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			imageURLs.Add(UnprocessedImageURL{imgURL})
			return "", false
		})
	}

	queue := NewImageQueue(imgFactory)

	resolvedImages, err := queue.Run(imageURLs, o.BuildConcurrency)
	if err != nil {
		return nil, err
	}

	return resolvedImages, nil
}

func (o *ResolveOptions) updateRefsInResources(nonConfigRs []ctlres.Resource,
	conf ctlconf.Conf, resolvedImages *ProcessedImages,
	imgFactory ctlimg.Factory) ([][]byte, error) {

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

func (o *ResolveOptions) emitLockOutput(conf ctlconf.Conf, resolvedImages *ProcessedImages) error {
	if len(o.LockOutput) == 0 {
		return nil
	}

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

	c.Overrides = ctlconf.UniqueImageOverrides(c.Overrides)

	return c.WriteToFile(o.LockOutput)
}
