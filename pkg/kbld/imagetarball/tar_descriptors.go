package tarball

import (
	"fmt"
	"io"
	"strings"
	"sync"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	regtran "github.com/google/go-containerregistry/pkg/v1/remote/transport"
	regtypes "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/k14s/kbld/pkg/kbld/util"
	"golang.org/x/sync/errgroup"
)

type TarDescriptorsMetadata interface {
	Generic(regname.Reference) (regv1.Descriptor, error)
	Index(regname.Reference) (regv1.ImageIndex, error)
	Image(regname.Reference) (regv1.Image, error)
}

type TarDescriptors struct {
	metadata TarDescriptorsMetadata

	tds []ImageOrImageIndexTarDescriptor

	imageLayersLock sync.Mutex
	imageLayers     map[ImageLayerTarDescriptor]regv1.Layer
}

func NewTarDescriptors(refs []regname.Reference, metadata TarDescriptorsMetadata) (*TarDescriptors, error) {
	metadata = errTarDescriptorsMetadata{metadata}

	tds := &TarDescriptors{
		metadata:    metadata,
		imageLayers: map[ImageLayerTarDescriptor]regv1.Layer{},
	}

	var tdsLock sync.Mutex
	var wg errgroup.Group
	buildThrottle := util.NewThrottle(10)

	for _, ref := range refs {
		ref := ref //copy

		wg.Go(func() error {
			buildThrottle.Take()
			defer buildThrottle.Done()

			desc, err := metadata.Generic(ref)
			if err != nil {
				return err
			}

			var td ImageOrImageIndexTarDescriptor

			if tds.isImageIndex(desc) {
				imgIndexTd, err := tds.buildImageIndex(ref, desc)
				if err != nil {
					return err
				}
				td = ImageOrImageIndexTarDescriptor{ImageIndex: &imgIndexTd}
			} else {
				imgTd, err := tds.buildImage(ref)
				if err != nil {
					return err
				}
				td = ImageOrImageIndexTarDescriptor{Image: &imgTd}
			}

			tdsLock.Lock()
			tds.tds = append(tds.tds, td)
			tdsLock.Unlock()

			return nil
		})
	}

	err := wg.Wait()

	return tds, err
}

func (tds *TarDescriptors) buildImageIndex(ref regname.Reference, desc regv1.Descriptor) (ImageIndexTarDescriptor, error) {
	td := ImageIndexTarDescriptor{
		Refs:      []string{ref.Name()},
		MediaType: string(desc.MediaType),
		Digest:    desc.Digest.String(),
	}

	imgIndex, err := tds.metadata.Index(ref)
	if err != nil {
		return td, err
	}

	rawManifest, err := imgIndex.RawManifest()
	if err != nil {
		return td, err
	}

	td.Raw = string(rawManifest)

	imgIndexManifest, err := imgIndex.IndexManifest()
	if err != nil {
		return td, err
	}

	for _, manDesc := range imgIndexManifest.Manifests {
		if tds.isImageIndex(manDesc) {
			imgIndexTd, err := tds.buildImageIndex(tds.buildRef(ref, manDesc.Digest.String()), manDesc)
			if err != nil {
				return ImageIndexTarDescriptor{}, err
			}
			td.Indexes = append(td.Indexes, imgIndexTd)
		} else {
			imgTd, err := tds.buildImage(tds.buildRef(ref, manDesc.Digest.String()))
			if err != nil {
				return ImageIndexTarDescriptor{}, err
			}
			td.Images = append(td.Images, imgTd)
		}
	}

	return td, nil
}

func (tds *TarDescriptors) buildImage(ref regname.Reference) (ImageTarDescriptor, error) {
	td := ImageTarDescriptor{}

	img, err := tds.metadata.Image(ref)
	if err != nil {
		return td, err
	}

	cfgDigest, err := img.ConfigName()
	if err != nil {
		return td, err
	}
	cfgBlob, err := img.RawConfigFile()
	if err != nil {
		return td, err
	}

	manifestMediaType, err := img.MediaType()
	if err != nil {
		return td, err
	}
	manifestDigest, err := img.Digest()
	if err != nil {
		return td, err
	}
	manifestBlob, err := img.RawManifest()
	if err != nil {
		return td, err
	}

	td = ImageTarDescriptor{
		Refs: []string{ref.String()},

		Config: ConfigTarDescriptor{
			Digest: cfgDigest.String(),
			Raw:    string(cfgBlob),
		},

		Manifest: ManifestTarDescriptor{
			MediaType: string(manifestMediaType),
			Digest:    manifestDigest.String(),
			Raw:       string(manifestBlob),
		},
	}

	layers, err := img.Layers()
	if err != nil {
		return td, err
	}

	for _, layer := range layers {
		layerMediaType, err := layer.MediaType()
		if err != nil {
			return td, err
		}
		layerDigest, err := layer.Digest()
		if err != nil {
			return td, err
		}
		layerDiffID, err := layer.DiffID()
		if err != nil {
			return td, err
		}
		layerSize, err := layer.Size()
		if err != nil {
			return td, err
		}

		layerTD := ImageLayerTarDescriptor{
			MediaType: string(layerMediaType),
			Digest:    layerDigest.String(),
			DiffID:    layerDiffID.String(),
			Size:      layerSize,
		}

		td.Layers = append(td.Layers, layerTD)

		tds.imageLayersLock.Lock()
		tds.imageLayers[layerTD] = layer
		tds.imageLayersLock.Unlock()
	}

	return td, nil
}

func (TarDescriptors) isImageIndex(desc regv1.Descriptor) bool {
	switch desc.MediaType {
	case regtypes.OCIImageIndex, regtypes.DockerManifestList:
		return true
	}
	return false
}

func (tds *TarDescriptors) ImageLayerStream(td ImageLayerTarDescriptor) (io.Reader, error) {
	tds.imageLayersLock.Lock()
	defer tds.imageLayersLock.Unlock()

	layer, found := tds.imageLayers[td]
	if !found {
		panic(fmt.Sprintf("Expected to find stream for %#v", td))
	}

	reader, err := layer.Compressed()
	if err != nil {
		return nil, fmt.Errorf("Getting compressed layer: %s", err)
	}
	return reader, nil
}

func (tds *TarDescriptors) buildRef(otherRef regname.Reference, digest string) regname.Reference {
	newRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", otherRef.Context().Name(), digest))
	if err != nil {
		panic(fmt.Sprintf("Building new ref"))
	}
	return newRef
}

type errTarDescriptorsMetadata struct {
	delegate TarDescriptorsMetadata
}

func (m errTarDescriptorsMetadata) Generic(ref regname.Reference) (regv1.Descriptor, error) {
	desc, err := m.delegate.Generic(ref)
	return desc, m.betterErr(ref, err)
}

func (m errTarDescriptorsMetadata) Index(ref regname.Reference) (regv1.ImageIndex, error) {
	idx, err := m.delegate.Index(ref)
	return idx, m.betterErr(ref, err)
}

func (m errTarDescriptorsMetadata) Image(ref regname.Reference) (regv1.Image, error) {
	img, err := m.delegate.Image(ref)
	return img, m.betterErr(ref, err)
}

func (m errTarDescriptorsMetadata) betterErr(ref regname.Reference, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), string(regtran.ManifestUnknownErrorCode)) {
			err = fmt.Errorf("Encountered an error most likely because this image is in Docker Registry v1 format; only v2 or OCI image format is supported (underlying error: %s)", err)
		}
		err = fmt.Errorf("Working with %s: %s", ref.Name(), err)
	}
	return err
}
