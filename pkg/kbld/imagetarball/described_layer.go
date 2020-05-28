package tarball

import (
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

type describedLayer struct {
	iltd     ImageLayerDescriptor
	contents LayerContents
}

var _ regv1.Layer = describedLayer{}

func (l describedLayer) Digest() (regv1.Hash, error) { return regv1.NewHash(l.iltd.Digest) }
func (l describedLayer) DiffID() (regv1.Hash, error) { return regv1.NewHash(l.iltd.DiffID) }

func (l describedLayer) Compressed() (io.ReadCloser, error) { return l.contents.Open() }

func (l describedLayer) Uncompressed() (io.ReadCloser, error) {
	rc, err := l.contents.Open()
	if err != nil {
		return nil, err
	}
	return v1util.GzipReadCloser(rc), nil
}

func (l describedLayer) Size() (int64, error) { return l.iltd.Size, nil }

func (l describedLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.iltd.MediaType), nil
}
