// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"sort"
)

type UnprocessedImageURL struct {
	URL string
}

type UnprocessedImageURLs struct {
	urls map[UnprocessedImageURL]struct{}
}

func NewUnprocessedImageURLs() *UnprocessedImageURLs {
	return &UnprocessedImageURLs{map[UnprocessedImageURL]struct{}{}}
}

func (i *UnprocessedImageURLs) Add(url UnprocessedImageURL) {
	i.urls[url] = struct{}{}
}

func (i *UnprocessedImageURLs) All() []UnprocessedImageURL {
	var result []UnprocessedImageURL
	for url := range i.urls {
		result = append(result, url)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].URL < result[j].URL
	})
	return result
}
