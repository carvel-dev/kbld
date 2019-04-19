package cmd

import (
	"fmt"
	"sync"

	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

type ImageBuildQueue struct {
	imgFactory ctlimg.Factory

	outputImages     map[string]string
	outputImagesLock sync.Mutex

	outputErrs     []error
	outputErrsLock sync.Mutex
}

func NewImageBuildQueue(imgFactory ctlimg.Factory) *ImageBuildQueue {
	return &ImageBuildQueue{imgFactory: imgFactory}
}

func (b *ImageBuildQueue) Run(inputImages map[string]struct{}, numWorkers int) (map[string]string, error) {
	b.outputImages = map[string]string{}
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

func (b *ImageBuildQueue) worker(workWg *sync.WaitGroup, queueCh <-chan string) {
	for img := range queueCh {
		b.work(workWg, img)
	}
}

func (b *ImageBuildQueue) work(workWg *sync.WaitGroup, img string) {
	defer workWg.Done()

	imgURL, err := b.imgFactory.New(img).URL()
	if err != nil {
		b.outputErrsLock.Lock()
		b.outputErrs = append(b.outputErrs, fmt.Errorf("Resolving image '%s': %s", img, err))
		b.outputErrsLock.Unlock()
		return
	}

	b.outputImagesLock.Lock()
	b.outputImages[img] = imgURL
	b.outputImagesLock.Unlock()
}
