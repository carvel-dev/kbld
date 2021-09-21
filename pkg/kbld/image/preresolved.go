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
	return PreresolvedImage{url, copyAndAppendMeta(metas)}
}

func (i PreresolvedImage) URL() (string, []ctlconf.Meta, error) {
	imageMetas := copyAndAppendMeta(i.metas, ctlconf.PreresolvedImageSourceURL{Type: ctlconf.PreresolvedMeta, URL: i.url})
	return i.url, imageMetas, nil
}

func copyAndAppendMeta(existing []ctlconf.Meta, new ...ctlconf.Meta) []ctlconf.Meta {
	all := make([]ctlconf.Meta, len(existing), len(existing)+len(new))
	copy(all, existing)
	return append(all, new...)
}
