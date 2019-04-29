package tarball

import (
	"fmt"
	"io"

	regname "github.com/google/go-containerregistry/pkg/name"
	regv1 "github.com/google/go-containerregistry/pkg/v1"
	regtypes "github.com/google/go-containerregistry/pkg/v1/types"
)

type TarDescriptorsMetadata interface {
	Generic(regname.Reference) (regv1.Descriptor, error)
	Index(regname.Reference) (regv1.ImageIndex, error)
	Image(regname.Reference) (regv1.Image, error)
}

type TarDescriptors struct {
	tds         []ImageOrImageIndexTarDescriptor
	imageLayers map[ImageLayerTarDescriptor]regv1.Layer
	metadata    TarDescriptorsMetadata
}

func NewTarDescriptors(refs []regname.Reference, metadata TarDescriptorsMetadata) (*TarDescriptors, error) {
	tds := &TarDescriptors{
		imageLayers: map[ImageLayerTarDescriptor]regv1.Layer{},
		metadata:    metadata,
	}

	for _, ref := range refs {
		desc, err := metadata.Generic(ref)
		if err != nil {
			return tds, err
		}

		var td ImageOrImageIndexTarDescriptor

		if tds.isImageIndex(desc) {
			imgIndexTd, err := tds.buildImageIndex(ref, desc)
			if err != nil {
				return nil, err
			}
			td = ImageOrImageIndexTarDescriptor{ImageIndex: &imgIndexTd}
		} else {
			imgTd, err := tds.buildImage(ref)
			if err != nil {
				return nil, err
			}
			td = ImageOrImageIndexTarDescriptor{Image: &imgTd}
		}

		tds.tds = append(tds.tds, td)
	}

	return tds, nil
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

		tds.imageLayers[layerTD] = layer
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
	layer, found := tds.imageLayers[td]
	if !found {
		panic(fmt.Sprintf("Expected to find stream for %#v", td))
	}
	return layer.Compressed()
}

func (tds *TarDescriptors) buildRef(otherRef regname.Reference, digest string) regname.Reference {
	newRef, err := regname.NewDigest(fmt.Sprintf("%s@%s", otherRef.Context().Name(), digest))
	if err != nil {
		panic(fmt.Sprintf("Building new ref"))
	}
	return newRef
}
