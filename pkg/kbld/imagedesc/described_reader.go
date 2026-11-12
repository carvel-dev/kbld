// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package imagedesc

import (
	"fmt"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

type DescribedReader struct {
	ids           *ImageRefDescriptors
	layerProvider LayerProvider
}

func NewDescribedReader(ids *ImageRefDescriptors, layerProvider LayerProvider) DescribedReader {
	return DescribedReader{ids, layerProvider}
}

func (r DescribedReader) Read() ([]ImageOrIndex, error) {
	var result []ImageOrIndex

	for i, td := range r.ids.Descriptors() {
		switch {
		case td.Image != nil:
			var img ImageWithRef = NewDescribedImage(*td.Image, r.layerProvider)
			result = append(result, ImageOrIndex{Image: &img})

		case td.ImageIndex != nil:
			idx := r.buildIndex(*td.ImageIndex)
			result = append(result, ImageOrIndex{Index: &idx})

		default:
			// Descriptors come from a manifest.json inside the tar, so an entry
			// carrying neither an image nor an index is untrusted input rather
			// than a broken invariant.
			return nil, fmt.Errorf("Expected descriptor %d to have either an image or an image index", i)
		}
	}

	return result, nil
}

func (r DescribedReader) buildIndex(iitd ImageIndexDescriptor) ImageIndexWithRef {
	var images []regv1.Image
	var indexes []regv1.ImageIndex

	for _, imgTD := range iitd.Images {
		images = append(images, NewDescribedImage(imgTD, r.layerProvider))
	}
	for _, indexTD := range iitd.Indexes {
		indexes = append(indexes, r.buildIndex(indexTD))
	}

	return NewDescribedImageIndex(iitd, images, indexes)
}
