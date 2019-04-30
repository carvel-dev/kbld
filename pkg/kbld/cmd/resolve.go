package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/ghodss/yaml"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
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
	BuildConcurrency int

	ExportImages     string
	ImportImages     string
	ImportRepository string
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
	cmd.Flags().IntVar(&o.BuildConcurrency, "build-concurrency", 4, "Set maximum number of concurrent builds")
	return cmd
}

func (o *ResolveOptions) Run() error {
	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("resolve | ")

	nonConfigRs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	resolvedImages, err := o.resolveImages(nonConfigRs, conf, logger)
	if err != nil {
		return err
	}

	// Record final image transformation
	for img, outputImg := range resolvedImages {
		prefixedLogger.WriteStr("final: %s -> %s\n", img, outputImg)
	}

	resBss, err := o.updateRefsInResources(nonConfigRs, resolvedImages)
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
	nonConfigRs []ctlres.Resource, conf ctlconf.Conf, logger ctlimg.Logger) (map[string]string, error) {

	foundImages := map[string]struct{}{}

	for _, res := range nonConfigRs {
		visitValues(res.DeepCopyRaw(), imageKey, func(val interface{}) (interface{}, bool) {
			if img, ok := val.(string); ok {
				foundImages[img] = struct{}{}
			}
			return nil, false
		})
	}

	queue := NewImageBuildQueue(ctlimg.NewFactory(conf, logger))

	resolvedImages, err := queue.Run(foundImages, o.BuildConcurrency)
	if err != nil {
		return nil, err
	}

	return resolvedImages, nil
}

func (o *ResolveOptions) updateRefsInResources(
	nonConfigRs []ctlres.Resource, resolvedImages map[string]string) ([][]byte, error) {

	var missingImageErrs []error
	var resBss [][]byte

	for _, res := range nonConfigRs {
		resContents := res.DeepCopyRaw()

		visitValues(resContents, imageKey, func(val interface{}) (interface{}, bool) {
			if img, ok := val.(string); ok {
				if outputImg, found := resolvedImages[img]; found {
					return outputImg, true
				}
				missingImageErrs = append(missingImageErrs, fmt.Errorf("Expected to find image for '%s'", img))
			}
			return nil, false
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
