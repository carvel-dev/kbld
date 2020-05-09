package cmd

import (
	"sort"
	"sync"
)

type ProcessedImages struct {
	imgs     map[UnprocessedImageURL]Image
	imgsLock sync.Mutex
}

type ProcessedImageItem struct {
	UnprocessedImageURL
	Image
}

func NewProcessedImages() *ProcessedImages {
	return &ProcessedImages{imgs: map[UnprocessedImageURL]Image{}}
}

func (i *ProcessedImages) Add(unprocessedImageURL UnprocessedImageURL, img Image) {
	i.imgsLock.Lock()
	defer i.imgsLock.Unlock()

	i.imgs[unprocessedImageURL] = img
}

func (i *ProcessedImages) FindByURL(unprocessedImageURL UnprocessedImageURL) (Image, bool) {
	i.imgsLock.Lock()
	defer i.imgsLock.Unlock()

	img, found := i.imgs[unprocessedImageURL]
	return img, found
}

func (i *ProcessedImages) All() []ProcessedImageItem {
	i.imgsLock.Lock()
	defer i.imgsLock.Unlock()

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
