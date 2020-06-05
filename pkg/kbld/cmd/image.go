package cmd

import (
	"strings"

	"github.com/ghodss/yaml"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

type Images []Image

type Image struct {
	URL      string
	Metas    []ctlimg.ImageMeta // empty when deserialized
	metasRaw []interface{}      // populated when deserialized
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

func (i Image) Metadata() []interface{} {
	var result []interface{}
	for _, meta := range i.Metas {
		result = append(result, meta)
	}

	return result
}

type imageStruct struct {
	URL   string
	Metas []interface{}
}

func newImageStructs(images []Image) []imageStruct {
	var result []imageStruct
	for _, img := range images {
		result = append(result, newImageStruct(img))
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
