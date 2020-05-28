package cmd

import (
	"fmt"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	regtarball "github.com/k14s/kbld/pkg/kbld/imagetarball"
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	"github.com/k14s/kbld/pkg/kbld/util"
)

type ImageSet struct {
	concurrency int
	logger      *ctlimg.LoggerPrefixWriter
}

func (o *ImageSet) Import(imgOrIndexes []regtarball.TarImageOrIndex,
	importRepo regname.Repository, registry ctlreg.Registry) (*ProcessedImages, error) {

	importedImages := NewProcessedImages()

	o.logger.WriteStr("importing %d images...\n", len(imgOrIndexes))
	defer func() { o.logger.WriteStr("imported %d images\n", len(importedImages.All())) }()

	errCh := make(chan error, len(imgOrIndexes))
	importThrottle := util.NewThrottle(o.concurrency)

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

			importDigestRef, err := o.importImage(item, existingRef, importRepo, registry)
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

func (o *ImageSet) importImage(item regtarball.TarImageOrIndex,
	existingRef regname.Digest, importRepo regname.Repository,
	registry ctlreg.Registry) (regname.Digest, error) {

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

	o.logger.Write([]byte(fmt.Sprintf("importing %s -> %s...\n", existingRef.Name(), importDigestRef.Name())))

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
	err = o.verifyTagDigest(uploadTagRef, importDigestRef, registry)
	if err != nil {
		return regname.Digest{}, err
	}

	return importDigestRef, nil
}

func (o *ImageSet) verifyTagDigest(
	uploadTagRef regname.Reference, importDigestRef regname.Digest, registry ctlreg.Registry) error {

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
