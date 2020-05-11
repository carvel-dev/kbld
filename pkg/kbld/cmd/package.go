package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	regtarball "github.com/k14s/kbld/pkg/kbld/imagetarball"
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

var _ regtarball.TarDescriptorsMetadata = ctlimg.Registry{}

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

	logger := ctlimg.NewLogger(os.Stderr)
	prefixedLogger := logger.NewPrefixedWriter("package | ")

	rs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := o.findImages(rs, conf, logger)
	if err != nil {
		return err
	}

	return o.exportImages(foundImages, prefixedLogger)
}

func (o *PackageOptions) findImages(allRs []ctlres.Resource,
	conf ctlconf.Conf, logger ctlimg.Logger) (*UnprocessedImageURLs, error) {

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

func (o *PackageOptions) exportImages(foundImages *UnprocessedImageURLs,
	logger *ctlimg.LoggerPrefixWriter) error {

	logger.WriteStr("exporting %d images...\n", len(foundImages.All()))
	defer func() { logger.WriteStr("exported %d images\n", len(foundImages.All())) }()

	var refs []regname.Reference

	for _, img := range foundImages.All() {
		// Validate strictly as these refs were already resolved
		ref, err := regname.NewDigest(img.URL, regname.StrictValidation)
		if err != nil {
			return err
		}

		logger.Write([]byte(fmt.Sprintf("will export %s\n", img.URL)))
		refs = append(refs, ref)
	}

	outputFile, err := os.Create(o.OutputPath)
	if err != nil {
		return fmt.Errorf("Creating file '%s': %s", o.OutputPath, err)
	}

	err = outputFile.Close()
	if err != nil {
		return err
	}

	outputFileOpener := func() (io.WriteCloser, error) {
		return os.OpenFile(o.OutputPath, os.O_RDWR, 0755)
	}

	tds, err := regtarball.NewTarDescriptors(refs, ctlimg.NewRegistry(o.RegistryFlags.AsRegistryOpts()))
	if err != nil {
		return fmt.Errorf("Collecting packaging metadata: %s", err)
	}

	logger.WriteStr("writing layers...\n")

	opts := regtarball.TarWriterOpts{Concurrency: o.Concurrency}

	return regtarball.NewTarWriter(tds, outputFileOpener, opts, logger).Write()
}
