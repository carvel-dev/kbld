// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"fmt"
	"strings"

	regname "github.com/google/go-containerregistry/pkg/name"
)

const digestSep = "@"

type DigestedImage struct {
	nameWithDigest regname.Digest
	parseErr       error
}

var _ Image = DigestedImage{}

func MaybeNewDigestedImage(url string) *DigestedImage {
	nameWithDigest, err := regname.NewDigest(url, regname.WeakValidation)
	if err != nil {
		if strings.Contains(url, digestSep) {
			return &DigestedImage{nameWithDigest, fmt.Errorf("Expected valid digest reference, but found '%s', reason: %s", url, err)}
		}
		return nil // not a reference with digest
	}
	return &DigestedImage{nameWithDigest, nil}
}

func NewDigestedImageFromParts(url, digest string) DigestedImage {
	ref := url + digestSep + digest

	nameWithDigest, err := regname.NewDigest(ref, regname.WeakValidation)
	if err != nil {
		return DigestedImage{nameWithDigest, fmt.Errorf("Expected digest reference, but found '%s', reason: %s", ref, err)}
	}
	return DigestedImage{nameWithDigest, nil}
}

func (i DigestedImage) URL() (string, []ImageMeta, error) {
	if i.parseErr != nil {
		return "", nil, i.parseErr
	}
	return i.nameWithDigest.Name(), nil, nil
}
