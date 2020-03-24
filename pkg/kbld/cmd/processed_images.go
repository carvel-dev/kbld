package cmd

import (
	"sort"
)

type ProcessedImages struct {
	imgs map[UnprocessedImageURL]Image
}

type ProcessedImageItem struct {
	UnprocessedImageURL
	Image
}

func NewProcessedImages() *ProcessedImages {
	return &ProcessedImages{map[UnprocessedImageURL]Image{}}
}

func (i *ProcessedImages) Add(unprocessedImageURL UnprocessedImageURL, img Image) {
	i.imgs[unprocessedImageURL] = img
}

func (i *ProcessedImages) FindByURL(unprocessedImageURL UnprocessedImageURL) (Image, bool) {
	img, found := i.imgs[unprocessedImageURL]
	return img, found
}

func (i *ProcessedImages) All() []ProcessedImageItem {
	var result []ProcessedImageItem
	for unprocessedImageURL, img := range i.imgs {
		result = append(result, ProcessedImageItem{
			UnprocessedImageURL: unprocessedImageURL,
			Image:               img,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UnprocessedImageURL.URL < result[j].UnprocessedImageURL.URL
	})
	return result
}
