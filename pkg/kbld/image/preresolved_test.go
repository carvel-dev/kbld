// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	. "carvel.dev/kbld/pkg/kbld/image"
	"github.com/stretchr/testify/assert"
)

func TestPreResolvedOrigins(t *testing.T) {
	preresolvedImage := "image-resolved-with-sha@sha256:9999"
	originsFromImagesLockConfig := []ctlconf.Origin{{Preresolved: &ctlconf.OriginPreresolved{URL: preresolvedImage}}}

	subject := NewPreresolvedImage(preresolvedImage, originsFromImagesLockConfig)

	url, origins, err := subject.URL()
	assert.NoError(t, err)
	assert.Equal(t, url, preresolvedImage)
	assert.Len(t, origins, 1)
}
