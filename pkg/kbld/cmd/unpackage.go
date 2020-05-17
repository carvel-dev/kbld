package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/ghodss/yaml"
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	regtarball "github.com/k14s/kbld/pkg/kbld/imagetarball"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
	"github.com/k14s/kbld/pkg/kbld/util"
	"github.com/k14s/kbld/pkg/kbld/version"
	"github.com/spf13/cobra"
)

type UnpackageOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
	InputPath     string
	Repository    string
	LockOutput    string
	Concurrency   int
}

func NewUnpackageOptions(ui ui.UI) *UnpackageOptions {
	return &UnpackageOptions{ui: ui}
}

func NewUnpackageCmd(o *UnpackageOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unpackage",
		Aliases: []string{"unpkg"},
		Short:   "Unpackage configuration and images from tarball",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	cmd.Flags().StringVarP(&o.InputPath, "input", "i", "", "Input tarball path")
	cmd.Flags().StringVarP(&o.Repository, "repository", "r", "", "Import images into given image repository (e.g. docker.io/dkalinin/my-project)")
	cmd.Flags().StringVar(&o.LockOutput, "lock-output", "", "File path to emit configuration with resolved image references")
	cmd.Flags().IntVar(&o.Concurrency, "concurrency", 5, "Set maximum number of concurrent imports")
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

	nonConfigRs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	// Import images used in the manifests
	importedImages, err := o.importImages(prefixedLogger)
	if err != nil {
		return err
	}

	err = o.emitLockOutput(conf, importedImages)
	if err != nil {
		return err
	}

	// Update previous image references with new references
	resBss, err := o.updateRefsInResources(nonConfigRs, conf, importedImages)
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

func (o *UnpackageOptions) updateRefsInResources(
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

func (o *UnpackageOptions) importImages(logger *ctlimg.LoggerPrefixWriter) (*ProcessedImages, error) {
	importedImages := NewProcessedImages()

	imgOrIndexes, err := regtarball.MultiRefReadFromFile(o.InputPath)
	if err != nil {
		return nil, err
	}

	logger.WriteStr("importing %d images...\n", len(imgOrIndexes))
	defer func() { logger.WriteStr("imported %d images\n", len(importedImages.All())) }()

	importRepo, err := regname.NewRepository(o.Repository)
	if err != nil {
		return nil, fmt.Errorf("Building import repository ref: %s", err)
	}

	errCh := make(chan error, len(imgOrIndexes))
	importThrottle := util.NewThrottle(o.Concurrency)

	for _, item := range imgOrIndexes {
		item := item // copy

		go func() {
			importThrottle.Take()
			defer importThrottle.Done()

			existingRef, err := regname.NewDigest(item.Ref())
			if err != nil {
				errCh <- err
				return
			}

			importDigestRef, err := o.importImage(item, existingRef, importRepo, logger)
			if err != nil {
				errCh <- fmt.Errorf("Importing image %s: %s", existingRef.Name(), err)
				return
			}

			importedImages.Add(UnprocessedImageURL{existingRef.Name()}, Image{URL: importDigestRef.Name()})
			errCh <- nil
		}()
	}

	for i := 0; i < len(imgOrIndexes); i++ {
		err := <-errCh
		if err != nil {
			return nil, err
		}
	}

	return importedImages, nil
}

func (o *UnpackageOptions) importImage(item regtarball.TarImageOrIndex,
	existingRef regname.Digest, importRepo regname.Repository,
	logger *ctlimg.LoggerPrefixWriter) (regname.Digest, error) {

	itemDigest, err := item.Digest()
	if err != nil {
		return regname.Digest{}, err
	}

	importDigestRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", importRepo.Name(), itemDigest))
	if err != nil {
		return regname.Digest{}, fmt.Errorf("Building new digest image ref: %s", err)
	}

	// Seems like AWS ECR doesnt like using digests for manifest uploads
	uploadTagRef, err := regname.NewTag(fmt.Sprintf("%s:kbld-%s-%s", importRepo.Name(), itemDigest.Algorithm, itemDigest.Hex))
	if err != nil {
		return regname.Digest{}, fmt.Errorf("Building upload tag image ref: %s", err)
	}

	logger.Write([]byte(fmt.Sprintf("importing %s -> %s...\n", existingRef.Name(), importDigestRef.Name())))

	registry := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())

	switch {
	case item.Image != nil:
		err = registry.WriteImage(uploadTagRef, *item.Image)
		if err != nil {
			return regname.Digest{}, fmt.Errorf("Importing image as %s: %s", importDigestRef.Name(), err)
		}

	case item.Index != nil:
		err = registry.WriteIndex(uploadTagRef, *item.Index)
		if err != nil {
			return regname.Digest{}, fmt.Errorf("Importing image index as %s: %s", importDigestRef.Name(), err)
		}

	default:
		panic("Unknown item")
	}

	// Verify that imported image still has the same digest as we expect.
	// Being a little bit paranoid here because tag ref is used for import
	// instead of plain digest ref, because AWS ECR doesnt like digests
	// during manifest upload.
	err = o.verifyTagDigest(uploadTagRef, importDigestRef)
	if err != nil {
		return regname.Digest{}, err
	}

	return importDigestRef, nil
}

func (o *UnpackageOptions) verifyTagDigest(
	uploadTagRef regname.Reference, importDigestRef regname.Digest) error {

	registry := ctlreg.NewRegistry(o.RegistryFlags.AsRegistryOpts())

	resultURL, _, err := ctlimg.NewResolvedImage(uploadTagRef.Name(), registry).URL()
	if err != nil {
		return fmt.Errorf("Verifying imported image %s: %s", uploadTagRef.Name(), err)
	}

	resultRef, err := regname.NewDigest(resultURL)
	if err != nil {
		return fmt.Errorf("Verifying imported image %s: %s", resultURL, err)
	}

	if resultRef.DigestStr() != importDigestRef.DigestStr() {
		return fmt.Errorf("Expected imported image '%s' to have digest '%s' but was '%s'",
			resultURL, importDigestRef.DigestStr(), resultRef.DigestStr())
	}

	return nil
}

func (o *UnpackageOptions) emitLockOutput(conf ctlconf.Conf, resolvedImages *ProcessedImages) error {
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
