package cmd

import (
	"fmt"
	"io"
	"os"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	"github.com/k14s/kbld/pkg/kbld/imagetar"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
)

type TarImageSet struct {
	imageSet    ImageSet
	concurrency int
	logger      *ctlimg.LoggerPrefixWriter
}

func (o TarImageSet) Export(foundImages *UnprocessedImageURLs,
	outputPath string, registry ctlreg.Registry) error {

	ids, err := o.imageSet.Export(foundImages, registry)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Creating file '%s': %s", outputPath, err)
	}

	err = outputFile.Close()
	if err != nil {
		return err
	}

	outputFileOpener := func() (io.WriteCloser, error) {
		return os.OpenFile(outputPath, os.O_RDWR, 0755)
	}

	o.logger.WriteStr("writing layers...\n")

	opts := imagetar.TarWriterOpts{Concurrency: o.concurrency}

	return imagetar.NewTarWriter(ids, outputFileOpener, opts, o.logger).Write()
}

func (o *TarImageSet) Import(path string,
	importRepo regname.Repository, registry ctlreg.Registry) (*ProcessedImages, error) {

	imgOrIndexes, err := imagetar.NewTarReader(path).Read()
	if err != nil {
		return nil, err
	}

	return o.imageSet.Import(imgOrIndexes, importRepo, registry)
}
