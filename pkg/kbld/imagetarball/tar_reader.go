package tarball

import (
	"io/ioutil"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

type TarReader struct {
	path string
}

func NewTarReader(path string) TarReader {
	return TarReader{path}
}

func (r TarReader) Read() ([]ImageOrIndex, error) {
	file := tarFile{r.path}

	manifestFile, err := file.Chunk("manifest.json").Open()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	ids, err := NewImageRefDescriptorsFromBytes(manifestBytes)
	if err != nil {
		return nil, err
	}

	return ReadFromTds(ids, file), nil
}

func ReadFromTds(ids *ImageRefDescriptors, layerProvider LayerProvider) []ImageOrIndex {
	var result []ImageOrIndex

	for _, td := range ids.descs {
		switch {
		case td.Image != nil:
			var img ImageWithRef = describedImage{*td.Image, layerProvider}
			result = append(result, ImageOrIndex{Image: &img})

		case td.ImageIndex != nil:
			idx := buildIndex(*td.ImageIndex, layerProvider)
			result = append(result, ImageOrIndex{Index: &idx})

		default:
			panic("Unknown item")
		}
	}

	return result
}

func buildIndex(iitd ImageIndexDescriptor, layerProvider LayerProvider) ImageIndexWithRef {
	var images []regv1.Image
	var indexes []regv1.ImageIndex

	for _, imgTD := range iitd.Images {
		images = append(images, describedImage{imgTD, layerProvider})
	}
	for _, indexTD := range iitd.Indexes {
		indexes = append(indexes, buildIndex(indexTD, layerProvider))
	}

	return describedImageIndex{iitd, images, indexes}
}
