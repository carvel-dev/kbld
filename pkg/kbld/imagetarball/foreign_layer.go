package tarball

import (
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type foreignLayer struct {
	iltd ImageLayerTarDescriptor
}

var _ regv1.Layer = foreignLayer{}

func (l foreignLayer) Digest() (regv1.Hash, error) { return regv1.NewHash(l.iltd.Digest) }
func (l foreignLayer) DiffID() (regv1.Hash, error) { return regv1.NewHash(l.iltd.DiffID) }

func (l foreignLayer) Compressed() (io.ReadCloser, error) {
	panic("foreignLayer.Compressed: not implemented")
}

func (l foreignLayer) Uncompressed() (io.ReadCloser, error) {
	panic("foreignLayer.Uncompressed: not implemented")
}

func (l foreignLayer) Size() (int64, error) { return l.iltd.Size, nil }

func (l foreignLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.iltd.MediaType), nil
}
