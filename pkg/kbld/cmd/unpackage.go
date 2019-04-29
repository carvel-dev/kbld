package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/ghodss/yaml"
	regname "github.com/google/go-containerregistry/pkg/name"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	regtarball "github.com/k14s/kbld/pkg/kbld/imagetarball"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type UnpackageOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags  FileFlags
	InputPath  string
	Repository string
}

func NewUnpackageOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *UnpackageOptions {
	return &UnpackageOptions{ui: ui, depsFactory: depsFactory}
}

func NewUnpackageCmd(o *UnpackageOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unpackage",
		Aliases: []string{"unpkg"},
		Short:   "Unpackage configuration and images from tarball",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	cmd.Flags().StringVarP(&o.InputPath, "input", "i", "", "Input tarball path")
	cmd.Flags().StringVarP(&o.Repository, "repository", "r", "", "Import images into given image repository (e.g. docker.io/dkalinin/my-project)")
	return cmd
}

func (o *UnpackageOptions) Run() error {
	if len(o.InputPath) == 0 {
		return fmt.Errorf("Expected 'input' flag to be non-empty")
	}
	if len(o.Repository) == 0 {
		return fmt.Errorf("Expected 'repository' flag to be non-empty")
	}

	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("unpackage | ")

	nonConfigRs, _, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	// Import images used in the manifests
	importedImages, err := o.importImages(prefixedLogger)
	if err != nil {
		return err
	}

	// Update previous image references with new references
	resBss, err := o.updateRefsInResources(nonConfigRs, importedImages)
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

func (o *UnpackageOptions) updateRefsInResources(
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

func (o *UnpackageOptions) importImages(logger *ctlimg.LoggerPrefixWriter) (map[string]string, error) {
	importedImages := map[string]string{}

	imgOrIndexes, err := regtarball.MultiRefReadFromFile(o.InputPath)
	if err != nil {
		return nil, err
	}

	logger.WriteStr("importing %d images...\n", len(imgOrIndexes))
	defer func() { logger.WriteStr("imported %d images\n", len(imgOrIndexes)) }()

	importRef, err := regname.NewRepository(o.Repository)
	if err != nil {
		return nil, fmt.Errorf("Building import repository ref: %s", err)
	}

	// TODO import in parallel?
	// TODO tag to avoid registry garbage collection?
	for _, item := range imgOrIndexes {
		existingRef, err := regname.NewDigest(item.Ref())
		if err != nil {
			return nil, err
		}

		itemDigest, err := item.Digest()
		if err != nil {
			return nil, err
		}

		newRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", importRef.Name(), itemDigest))
		if err != nil {
			return nil, fmt.Errorf("Building new image ref: %s", err)
		}

		logger.Write([]byte(fmt.Sprintf("importing %s -> %s...\n", existingRef.Name(), newRef.Name())))

		switch {
		case item.Image != nil:
			err = ctlimg.ResolvedImage{}.Write(newRef, *item.Image)
			if err != nil {
				return nil, fmt.Errorf("Importing image as %s: %s", newRef.Name(), err)
			}

		case item.Index != nil:
			err = ctlimg.ResolvedImage{}.WriteIndex(newRef, *item.Index)
			if err != nil {
				return nil, fmt.Errorf("Importing image index as %s: %s", newRef.Name(), err)
			}

		default:
			panic("Unknown item")
		}

		importedImages[existingRef.Name()] = newRef.Name()
	}

	return importedImages, nil
}
