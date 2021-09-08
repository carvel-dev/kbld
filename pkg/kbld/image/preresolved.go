// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
)

type PreresolvedImage struct {
	url   string
	metas []ctlconf.Meta
}

func NewPreresolvedImage(url string, metas []ctlconf.Meta) PreresolvedImage {
	return PreresolvedImage{url, metas}
}

func (i PreresolvedImage) URL() (string, []ctlconf.Meta, error) {
	imageMetas := append(i.metas, ctlconf.PreresolvedImageSourceURL{Type: "preresolved", URL: i.url})

	return i.url, imageMetas, nil
}
