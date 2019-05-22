package cmd

import (
	"fmt"
	"sync"

	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

type ImageQueue struct {
	imgFactory ctlimg.Factory

	outputImages     map[string]Image
	outputImagesLock sync.Mutex

	outputErrs     []error
	outputErrsLock sync.Mutex
}

func NewImageQueue(imgFactory ctlimg.Factory) *ImageQueue {
	return &ImageQueue{imgFactory: imgFactory}
}

func (b *ImageQueue) Run(inputImages map[string]struct{}, numWorkers int) (map[string]Image, error) {
	b.outputImages = map[string]Image{}
	b.outputErrs = nil

	queueCh := make(chan string, numWorkers)
	workWg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		go b.worker(&workWg, queueCh)
	}

	for img, _ := range inputImages {
		workWg.Add(1)
		queueCh <- img
	}

	workWg.Wait()
	close(queueCh)

	return b.outputImages, errFromErrs(b.outputErrs)
}

func (b *ImageQueue) worker(workWg *sync.WaitGroup, queueCh <-chan string) {
	for img := range queueCh {
		b.work(workWg, img)
	}
}

func (b *ImageQueue) work(workWg *sync.WaitGroup, img string) {
	defer workWg.Done()

	imgURL, metas, err := b.imgFactory.New(img).URL()
	if err != nil {
		b.outputErrsLock.Lock()
		b.outputErrs = append(b.outputErrs, fmt.Errorf("Resolving image '%s': %s", img, err))
		b.outputErrsLock.Unlock()
		return
	}

	b.outputImagesLock.Lock()
	b.outputImages[img] = Image{URL: imgURL, Metas: metas}
	b.outputImagesLock.Unlock()
}
