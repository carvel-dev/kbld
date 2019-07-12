package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

const (
	imageKey = "image"
)

type ResolveOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags        FileFlags
	RegistryFlags    RegistryFlags
	BuildConcurrency int
	ImagesAnnotation bool
}

func NewResolveOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ResolveOptions {
	return &ResolveOptions{ui: ui, depsFactory: depsFactory}
}

func NewResolveCmd(o *ResolveOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Build images and update references",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().IntVar(&o.BuildConcurrency, "build-concurrency", 4, "Set maximum number of concurrent builds")
	cmd.Flags().BoolVar(&o.ImagesAnnotation, "images-annotation", true, "Annotate resources with images annotation")
	return cmd
}

func (o *ResolveOptions) Run() error {
	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("resolve | ")

	nonConfigRs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	registry := ctlimg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	imgFactory := ctlimg.NewFactory(conf, registry, logger)

	resolvedImages, err := o.resolveImages(nonConfigRs, imgFactory)
	if err != nil {
		return err
	}

	// Record final image transformation
	for imgURL, img := range resolvedImages {
		prefixedLogger.WriteStr("final: %s -> %s\n", imgURL, img.URL)
	}

	resBss, err := o.updateRefsInResources(nonConfigRs, resolvedImages, imgFactory)
	if err != nil {
		return err
	}

	// Print all resources as one YAML stream
	for i, resBs := range resBss {
		if i != 0 {
			resBs = append([]byte("---\n"), resBs...)
		}
		o.ui.PrintBlock(resBs)
	}

	return nil
}

func (o *ResolveOptions) resolveImages(
	nonConfigRs []ctlres.Resource, imgFactory ctlimg.Factory) (map[string]Image, error) {

	foundImages := map[string]struct{}{}

	for _, res := range nonConfigRs {
		visitValues(res.DeepCopyRaw(), imageKey, func(val interface{}) (interface{}, bool) {
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
	resolvedImages map[string]Image, imgFactory ctlimg.Factory) ([][]byte, error) {

	var errs []error
	var resBss [][]byte

	for _, res := range nonConfigRs {
		resContents := res.DeepCopyRaw()
		images := []Image{}

		visitValues(resContents, imageKey, func(val interface{}) (interface{}, bool) {
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

func visitValues(obj interface{}, key string, visitorFunc func(interface{}) (interface{}, bool)) {
	switch typedObj := obj.(type) {
	case map[string]interface{}:
		for k, v := range typedObj {
			if k == key {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal
				}
			} else {
				visitValues(typedObj[k], key, visitorFunc)
			}
		}

	case map[string]string:
		for k, v := range typedObj {
			if k == key {
				if newVal, update := visitorFunc(v); update {
					typedObj[k] = newVal.(string)
				}
			} else {
				visitValues(typedObj[k], key, visitorFunc)
			}
		}

	case []interface{}:
		for _, o := range typedObj {
			visitValues(o, key, visitorFunc)
		}
	}
}
