package imagetar

import (
	"io/ioutil"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/k14s/kbld/pkg/kbld/imagedesc"
)

type TarReader struct {
	path string
}

func NewTarReader(path string) TarReader {
	return TarReader{path}
}

func (r TarReader) Read() ([]imagedesc.ImageOrIndex, error) {
	file := tarFile{r.path}

	manifestFile, err := file.Chunk("manifest.json").Open()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	ids, err := imagedesc.NewImageRefDescriptorsFromBytes(manifestBytes)
	if err != nil {
		return nil, err
	}

	return ReadFromTds(ids, file), nil
}

func ReadFromTds(ids *imagedesc.ImageRefDescriptors, layerProvider imagedesc.LayerProvider) []imagedesc.ImageOrIndex {
	var result []imagedesc.ImageOrIndex

	for _, td := range ids.Descriptors() {
		switch {
		case td.Image != nil:
			var img imagedesc.ImageWithRef = imagedesc.NewDescribedImage(*td.Image, layerProvider)
			result = append(result, imagedesc.ImageOrIndex{Image: &img})

		case td.ImageIndex != nil:
			idx := buildIndex(*td.ImageIndex, layerProvider)
			result = append(result, imagedesc.ImageOrIndex{Index: &idx})

		default:
			panic("Unknown item")
		}
	}

	return result
}

func buildIndex(iitd imagedesc.ImageIndexDescriptor,
	layerProvider imagedesc.LayerProvider) imagedesc.ImageIndexWithRef {

	var images []regv1.Image
	var indexes []regv1.ImageIndex

	for _, imgTD := range iitd.Images {
		images = append(images, imagedesc.NewDescribedImage(imgTD, layerProvider))
	}
	for _, indexTD := range iitd.Indexes {
		indexes = append(indexes, buildIndex(indexTD, layerProvider))
	}

	return imagedesc.NewDescribedImageIndex(iitd, images, indexes)
}
