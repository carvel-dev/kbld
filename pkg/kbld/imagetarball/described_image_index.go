package tarball

import (
	"encoding/json"
	"fmt"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type describedImageIndex struct {
	iitd    ImageIndexDescriptor
	images  []regv1.Image
	indexes []regv1.ImageIndex
}

var _ regv1.ImageIndex = describedImageIndex{}

func (i describedImageIndex) Ref() string { return i.iitd.Refs[0] }

func (i describedImageIndex) MediaType() (types.MediaType, error) {
	return types.MediaType(i.iitd.MediaType), nil
}

func (i describedImageIndex) Digest() (regv1.Hash, error)  { return regv1.NewHash(i.iitd.Digest) }
func (i describedImageIndex) RawManifest() ([]byte, error) { return []byte(i.iitd.Raw), nil }

func (i describedImageIndex) Size() (int64, error) { return int64(len(i.iitd.Raw)), nil }

func (i describedImageIndex) IndexManifest() (*regv1.IndexManifest, error) {
	var manifest *regv1.IndexManifest
	err := json.Unmarshal([]byte(i.iitd.Raw), &manifest)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (i describedImageIndex) Image(digest regv1.Hash) (regv1.Image, error) {
	for _, img := range i.images {
		imgDigest, err := img.Digest()
		if err != nil {
			return nil, err
		}
		if imgDigest.String() == digest.String() {
			return img, nil
		}
	}
	return nil, fmt.Errorf("Expected to find image '%s' by digest", digest)
}

func (i describedImageIndex) ImageIndex(digest regv1.Hash) (regv1.ImageIndex, error) {
	for _, idx := range i.indexes {
		idxDigest, err := idx.Digest()
		if err != nil {
			return nil, err
		}
		if idxDigest.String() == digest.String() {
			return idx, nil
		}
	}
	return nil, fmt.Errorf("Expected to find index '%s' by digest", digest)
}
