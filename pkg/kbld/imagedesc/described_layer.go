package imagedesc

import (
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

type DescribedLayer struct {
	desc     ImageLayerDescriptor
	contents LayerContents
}

var _ regv1.Layer = DescribedLayer{}

func NewDescribedLayer(desc ImageLayerDescriptor, contents LayerContents) DescribedLayer {
	return DescribedLayer{desc, contents}
}

func (l DescribedLayer) Digest() (regv1.Hash, error) { return regv1.NewHash(l.desc.Digest) }
func (l DescribedLayer) DiffID() (regv1.Hash, error) { return regv1.NewHash(l.desc.DiffID) }

func (l DescribedLayer) Compressed() (io.ReadCloser, error) { return l.contents.Open() }

func (l DescribedLayer) Uncompressed() (io.ReadCloser, error) {
	rc, err := l.contents.Open()
	if err != nil {
		return nil, err
	}
	return v1util.GzipReadCloser(rc), nil
}

func (l DescribedLayer) Size() (int64, error) { return l.desc.Size, nil }

func (l DescribedLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.desc.MediaType), nil
}
