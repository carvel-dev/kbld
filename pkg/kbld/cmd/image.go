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

func (i Image) Equal(other Image) bool {
	if i.URL != other.URL {
		return false
	}
	if len(i.Metas) != len(other.Metas) {
		return false
	}
	for i, meta := range i.Metas {
		if meta != other.Metas[i] {
			return false
		}
	}
	return true
}

func (imgs Images) Contains(img Image) bool {
	for _, other := range imgs {
		if img.Equal(other) {
			return true
		}
	}
	return false
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
