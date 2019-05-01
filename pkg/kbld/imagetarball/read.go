package tarball

import (
	"encoding/json"
	"io/ioutil"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

type ImageWithRef interface {
	regv1.Image
	Ref() string
}

type ImageIndexWithRef interface {
	regv1.ImageIndex
	Ref() string
}

type TarImageOrIndex struct {
	Image *ImageWithRef
	Index *ImageIndexWithRef
}

func (t TarImageOrIndex) Digest() (regv1.Hash, error) {
	switch {
	case t.Image != nil:
		return (*t.Image).Digest()
	case t.Index != nil:
		return (*t.Index).Digest()
	default:
		panic("Unknown item")
	}
}

func (t TarImageOrIndex) Ref() string {
	switch {
	case t.Image != nil:
		return (*t.Image).Ref()
	case t.Index != nil:
		return (*t.Index).Ref()
	default:
		panic("Unknown item")
	}
}

func MultiRefReadFromFile(path string) ([]TarImageOrIndex, error) {
	file := tarFile{path}

	manifestFile, err := file.Chunk("manifest.json").Open()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	var tds []ImageOrImageIndexTarDescriptor
	var result []TarImageOrIndex

	err = json.Unmarshal(manifestBytes, &tds)
	if err != nil {
		return nil, err
	}

	for _, td := range tds {
		switch {
		case td.Image != nil:
			var img ImageWithRef = tarImage{*td.Image, file}
			result = append(result, TarImageOrIndex{Image: &img})

		case td.ImageIndex != nil:
			idx := buildIndex(*td.ImageIndex, file)
			result = append(result, TarImageOrIndex{Index: &idx})

		default:
			panic("Unknown item")
		}
	}

	return result, nil
}

func buildIndex(iitd ImageIndexTarDescriptor, file tarFile) ImageIndexWithRef {
	var images []regv1.Image
	var indexes []regv1.ImageIndex

	for _, imgTD := range iitd.Images {
		images = append(images, tarImage{imgTD, file})
	}
	for _, indexTD := range iitd.Indexes {
		indexes = append(indexes, buildIndex(indexTD, file))
	}

	return tarImageIndex{iitd, images, indexes}
}
