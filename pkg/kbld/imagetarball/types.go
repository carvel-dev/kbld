package tarball

import (
	regv1types "github.com/google/go-containerregistry/pkg/v1/types"
)

type ImageOrImageIndexTarDescriptor struct {
	ImageIndex *ImageIndexTarDescriptor
	Image      *ImageTarDescriptor
}

type ImageIndexTarDescriptor struct {
	Refs    []string
	Images  []ImageTarDescriptor
	Indexes []ImageIndexTarDescriptor

	MediaType string
	Digest    string
	Raw       string
}

type ImageTarDescriptor struct {
	Refs   []string
	Layers []ImageLayerTarDescriptor

	Config   ConfigTarDescriptor
	Manifest ManifestTarDescriptor
}

type ImageLayerTarDescriptor struct {
	MediaType string
	Digest    string
	DiffID    string
	Size      int64
}

type ConfigTarDescriptor struct {
	Digest string
	Raw    string
}

type ManifestTarDescriptor struct {
	MediaType string
	Digest    string
	Raw       string
}

func (td ImageLayerTarDescriptor) IsDistributable() bool {
	// Example layer representation for windows rootfs:
	//   "mediaType": "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
	//   "size": 1654613376,
	//   "digest": "sha256:31f9df80631e7b5d379647ee7701ff50e009bd2c03b30a67a0a8e7bba4a26f94",
	//   "urls": ["https://mcr.microsoft.com/v2/windows/servercore/blobs/sha256:31f9df80631e7b5d379647ee7701ff50e009bd2c03b30a67a0a8e7bba4a26f94"]
	return regv1types.MediaType(td.MediaType).IsDistributable()
}
