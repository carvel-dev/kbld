// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"sort"

	"sigs.k8s.io/yaml"
)

type UnprocessedImageURL struct {
	URL string `json:"image"`
}

type UnprocessedImageURLs struct {
	urls map[UnprocessedImageURL]struct{} `json:"unresolved"`
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

func (i *UnprocessedImageURLs) Bytes() ([]byte, error) {
	return yaml.Marshal(i.All())
}
