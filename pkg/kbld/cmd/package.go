package cmd

import (
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	regtarball "github.com/k14s/kbld/pkg/kbld/imagetarball"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type PackageOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
	OutputPath    string
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
	conf ctlconf.Conf, logger ctlimg.Logger) (map[string]struct{}, error) {

	foundImages := map[string]struct{}{}

	for _, res := range allRs {
		imageKVs := ImageKVs{res.DeepCopyRaw(), conf.ImageKeys()}

		imageKVs.Visit(func(val interface{}) (interface{}, bool) {
			if img, ok := val.(string); ok {
				foundImages[img] = struct{}{}
			}
			return nil, false
		})
	}

	for img, _ := range foundImages {
		_, err := regname.NewDigest(img)
		if err != nil {
			return nil, fmt.Errorf("Expected image '%s' to be in digest form (i.e. image@digest)", img)
		}
	}

	return foundImages, nil
}

func (o *PackageOptions) exportImages(imgRefsToExport map[string]struct{}, logger *ctlimg.LoggerPrefixWriter) error {
	logger.WriteStr("exporting %d images...\n", len(imgRefsToExport))
	defer func() { logger.WriteStr("exported %d images\n", len(imgRefsToExport)) }()

	var refs []regname.Reference

	for imgRef, _ := range imgRefsToExport {
		// Validate strictly as these refs were already resolved
		ref, err := regname.NewDigest(imgRef, regname.StrictValidation)
		if err != nil {
			return err
		}

		logger.Write([]byte(fmt.Sprintf("will export %s\n", imgRef)))
		refs = append(refs, ref)
	}

	outputFile, err := os.Create(o.OutputPath)
	if err != nil {
		return err
	}

	defer outputFile.Close()

	tds, err := regtarball.NewTarDescriptors(refs, ctlimg.NewRegistry(o.RegistryFlags.AsRegistryOpts()))
	if err != nil {
		return fmt.Errorf("Collecting packaging metadata: %s", err)
	}

	return regtarball.NewTarWriter(tds, outputFile).Write()
}
