package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/ghodss/yaml"
	regname "github.com/google/go-containerregistry/pkg/name"
	regtypes "github.com/google/go-containerregistry/pkg/v1/types"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	"github.com/k14s/kbld/pkg/kbld/image"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
	"github.com/spf13/cobra"
)

type RelocateOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
	Repository    string
	// concurrency?

}

func NewRelocateOptions(ui ui.UI) *RelocateOptions {
	return &RelocateOptions{ui: ui}
}

func NewRelocateCmd(o *RelocateOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relocate",
		Short: "Relocate images between two registries",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}

	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().StringVarP(&o.Repository, "repository", "r", "", "Import images into given image repository (e.g. docker.io/dkalinin/my-project)")

	return cmd
}

func (o *RelocateOptions) Run() error {
	// basic checks
	if len(o.Repository) == 0 {
		return fmt.Errorf("Expected repository flag to be non-empty")
	}

	if len(o.FileFlags.Files) == 0 {
		return fmt.Errorf("Expected at least one input file")
	}

	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("relocate | ")

	// get resources from files
	rs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := o.findImages(rs, conf)
	if err != nil {
		return err
	}

	importedImages, err := o.relocateImages(foundImages, prefixedLogger)
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

func (o *RelocateOptions) relocateImages(foundImages *UnprocessedImageURLs,
	logger *image.LoggerPrefixWriter) (*ProcessedImages, error) {

	importedImages := NewProcessedImages()
	// create registries
	srcRegistry := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())
	// may need dst registry flags
	dstRegistry := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())

	dstRepo, err := regname.NewRepository(o.Repository)
	if err != nil {
		return nil, err
	}

	for _, img := range foundImages.All() {
		srcRef, err := regname.NewDigest(img.URL, regname.StrictValidation)
		if err != nil {
			return nil, fmt.Errorf("Error building source ref: %s", err)
		}

		desc, err := srcRegistry.Generic(srcRef)
		if err != nil {
			return nil, err
		}
		switch desc.MediaType {
		case regtypes.OCIImageIndex, regtypes.DockerManifestList:
			idx, err := srcRegistry.Index(srcRef)
			if err != nil {
				return nil, fmt.Errorf("Error getting remote image: %s", err)
			}

			idxDigest, err := idx.Digest()
			if err != nil {
				return nil, err
			}

			dstRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", dstRepo.Name(), idxDigest))
			if err != nil {
				return nil, fmt.Errorf("Error building destination ref: %s", err)
			}

			logger.Write([]byte(fmt.Sprintf("importing %s -> %s...\n", srcRef.Name(), dstRef.Name())))

			err = dstRegistry.WriteIndex(dstRef, idx)
			if err != nil {
				return nil, err
			}
			importedImages.Add(UnprocessedImageURL{srcRef.Name()}, Image{URL: dstRef.Name()})

		default:
			img, err := srcRegistry.Image(srcRef)
			if err != nil {
				return nil, fmt.Errorf("Error getting remote image: %s", err)
			}

			imgDigest, err := img.Digest()
			if err != nil {
				return nil, err
			}

			dstRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", dstRepo.Name(), imgDigest))
			if err != nil {
				return nil, fmt.Errorf("Error building destination ref: %s", err)
			}

			logger.Write([]byte(fmt.Sprintf("importing %s -> %s...\n", srcRef.Name(), dstRef.Name())))

			err = dstRegistry.WriteImage(dstRef, img)
			if err != nil {
				return nil, err
			}
			importedImages.Add(UnprocessedImageURL{srcRef.Name()}, Image{URL: dstRef.Name()})
		}
	}

	return importedImages, nil
}

func (o *RelocateOptions) findImages(allRs []ctlres.Resource,
	conf ctlconf.Conf) (*UnprocessedImageURLs, error) {

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
