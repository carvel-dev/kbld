// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

type PreresolvedImage struct {
	url string
}

type PreresolvedImageSourceURL struct {
	Type string // always set to 'preresolved'
	URL  string
}

func (PreresolvedImageSourceURL) meta() {}

func NewPreresolvedImage(url string) PreresolvedImage {
	return PreresolvedImage{url}
}

func (i PreresolvedImage) URL() (string, []Meta, error) {
	return i.url, []Meta{PreresolvedImageSourceURL{Type: "preresolved", URL: i.url}}, nil
}
