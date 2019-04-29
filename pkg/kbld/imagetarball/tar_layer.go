package tarball

import (
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/go-containerregistry/pkg/v1/v1util"
)

type tarLayer struct {
	iltd      ImageLayerTarDescriptor
	fileChunk tarFileChunk
}

var _ regv1.Layer = tarLayer{}

func (l tarLayer) Digest() (regv1.Hash, error) { return regv1.NewHash(l.iltd.Digest) }
func (l tarLayer) DiffID() (regv1.Hash, error) { return regv1.NewHash(l.iltd.DiffID) }

func (l tarLayer) Compressed() (io.ReadCloser, error) { return l.fileChunk.Open() }

func (l tarLayer) Uncompressed() (io.ReadCloser, error) {
	rc, err := l.fileChunk.Open()
	if err != nil {
		return nil, err
	}
	return v1util.GzipReadCloser(rc)
}

func (l tarLayer) Size() (int64, error) { return l.iltd.Size, nil }

func (l tarLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.iltd.MediaType), nil
}
