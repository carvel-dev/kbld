// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package imagedesc

import (
	"io"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type ForeignDescribedLayer struct {
	desc ImageLayerDescriptor
}

var _ regv1.Layer = ForeignDescribedLayer{}

func NewForeignDescribedLayer(desc ImageLayerDescriptor) ForeignDescribedLayer {
	return ForeignDescribedLayer{desc}
}

func (l ForeignDescribedLayer) Digest() (regv1.Hash, error) { return regv1.NewHash(l.desc.Digest) }
func (l ForeignDescribedLayer) DiffID() (regv1.Hash, error) { return regv1.NewHash(l.desc.DiffID) }

func (l ForeignDescribedLayer) Compressed() (io.ReadCloser, error) {
	panic("ForeignDescribedLayer.Compressed: not implemented")
}

func (l ForeignDescribedLayer) Uncompressed() (io.ReadCloser, error) {
	panic("ForeignDescribedLayer.Uncompressed: not implemented")
}

func (l ForeignDescribedLayer) Size() (int64, error) { return l.desc.Size, nil }

func (l ForeignDescribedLayer) MediaType() (types.MediaType, error) {
	return types.MediaType(l.desc.MediaType), nil
}
