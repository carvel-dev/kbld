// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"reflect"
	"strings"

	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	"sigs.k8s.io/yaml"
)

type Images []Image

type Image struct {
	URL      string
	Metas    []ctlconf.Meta // empty when deserialized
	metasRaw []interface{} // populated when deserialized
}

func (imgs Images) ForImage(url string) (Image, bool) {
	for _, img := range imgs {
		if img.URL == url {
			return img, true
		}
	}
	return Image{}, false
}

// TODO only works after deserialization
func (i Image) Description() string {
	yamlBytes, err := yaml.Marshal(i.metasRaw)
	if err != nil {
		return "[]" // TODO deal better?
	}

	return strings.TrimSpace(string(yamlBytes))
}

type imageStruct struct {
	URL   string
	Metas []interface{}
}

func (st imageStruct) equal(other imageStruct) bool {
	return st.URL == other.URL && reflect.DeepEqual(st.Metas, other.Metas)
}

func contains(structs []imageStruct, st imageStruct) bool {
	for _, other := range structs {
		if st.equal(other) {
			return true
		}
	}
	return false
}

func newImageStructs(images []Image) []imageStruct {
	var result []imageStruct
	for _, img := range images {
		st := newImageStruct(img)
		// if Metas is empty then the image was already in digest form and we didn't need to resolve
		// it, so the annotation isn't very useful
		if len(st.Metas) > 0 {
			// also check for duplicates before adding
			if !contains(result, st) {
				result = append(result, st)
			}
		}
	}
	return result
}

func newImageStruct(image Image) imageStruct {
	result := imageStruct{URL: image.URL}
	for _, meta := range image.Metas {
		result.Metas = append(result.Metas, meta)
	}
	return result
}

func newImages(structs []imageStruct) []Image {
	var result []Image
	for _, st := range structs {
		result = append(result, Image{URL: st.URL, metasRaw: st.Metas})
	}
	return result
}
