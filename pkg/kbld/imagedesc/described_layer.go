// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package imagedesc

import (
	"fmt"
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/k14s/kbld/pkg/kbld/imageutils/gzip"
	"github.com/k14s/kbld/pkg/kbld/imageutils/verify"
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

	h, err := l.Digest()
	if err != nil {
		return nil, fmt.Errorf("Computing digest: %v", err)
	}

	rc, err = verify.ReadCloser(rc, h)
	if err != nil {
		return nil, fmt.Errorf("Creating verified reader: %v", err)
	}

	return gzip.ReadCloser(rc), nil
}

func (l DescribedLayer) Size() (int64, error) { return l.desc.Size, nil }

func (l DescribedLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.desc.MediaType), nil
}
