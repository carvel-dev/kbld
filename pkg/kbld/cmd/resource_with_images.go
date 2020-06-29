package cmd

import (
	"sort"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ImagesAnnKey = "kbld.k14s.io/images"
)

type ResourceWithImages struct {
	contents map[string]interface{}
	images   []Image
}

func NewResourceWithImages(contents map[string]interface{}, images []Image) ResourceWithImages {
	// sort images lexicographically (by URL) to avoid unnecessary annotation changes
	sort.Slice(images, func(i, j int) bool {
		return images[i].URL < images[j].URL
	})
	return ResourceWithImages{contents, images}
}

func (r ResourceWithImages) Bytes() ([]byte, error) {
	if len(r.images) > 0 {
		resUn := unstructured.Unstructured{r.contents}

		imagesYAML, err := yaml.Marshal(newImageStructs(r.images))
		if err != nil {
			return nil, err
		}

		anns := resUn.GetAnnotations()
		if anns == nil {
			anns = map[string]string{}
		}

		anns[ImagesAnnKey] = string(imagesYAML)
		resUn.SetAnnotations(anns)
		r.contents = resUn.Object
	}

	return yaml.Marshal(r.contents)
}

func (r ResourceWithImages) Images() ([]Image, error) {
	resUn := unstructured.Unstructured{r.contents}

	anns := resUn.GetAnnotations()
	if anns == nil {
		return nil, nil
	}

	var structs []imageStruct

	err := yaml.Unmarshal([]byte(anns[ImagesAnnKey]), &structs)
	if err != nil {
		return nil, err
	}

	return newImages(structs), nil
}
