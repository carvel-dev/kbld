package cmd

import (
	"fmt"

	"github.com/ghodss/yaml"
	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	SourcesAnnKey = "kbld.k14s.io/sources"
)

type ResourceWithBuiltImages struct {
	contents map[string]interface{}
	metas    []BuiltImageMeta
}

type BuiltImageMeta struct {
	Image string
	ctlimg.BuiltImageSources
}

type BuiltImageMetas []BuiltImageMeta

func NewResourceWithBuiltImages(contents map[string]interface{}, metas []BuiltImageMeta) ResourceWithBuiltImages {
	return ResourceWithBuiltImages{contents, metas}
}

func (r ResourceWithBuiltImages) Bytes() ([]byte, error) {
	if len(r.metas) > 0 {
		resUn := unstructured.Unstructured{r.contents}

		metasYAML, err := yaml.Marshal(r.metas)
		if err != nil {
			return nil, err
		}

		anns := resUn.GetAnnotations()
		if anns == nil {
			anns = map[string]string{}
		}

		anns[SourcesAnnKey] = string(metasYAML)
		resUn.SetAnnotations(anns)
		r.contents = resUn.Object
	}

	return yaml.Marshal(r.contents)
}

func (r ResourceWithBuiltImages) Metas() ([]BuiltImageMeta, error) {
	resUn := unstructured.Unstructured{r.contents}

	anns := resUn.GetAnnotations()
	if anns == nil {
		return nil, nil
	}

	var metas []BuiltImageMeta

	err := yaml.Unmarshal([]byte(anns[SourcesAnnKey]), &metas)
	if err != nil {
		return nil, err
	}

	return metas, nil
}

func (ms BuiltImageMetas) ForImage(url string) (BuiltImageMeta, error) {
	for _, meta := range ms {
		if meta.Image == url {
			return meta, nil
		}
	}
	return BuiltImageMeta{}, fmt.Errorf("Expected to find meta for image '%s'", url)
}
