// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"sync"

	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

type ImageQueue struct {
	imgFactory ctlimg.Factory

	outputImages     *ProcessedImages
	outputImagesLock sync.Mutex

	outputErrs     []error
	outputErrsLock sync.Mutex
}

func NewImageQueue(imgFactory ctlimg.Factory) *ImageQueue {
	return &ImageQueue{imgFactory: imgFactory}
}

func (b *ImageQueue) Run(unprocessedImageURLs *UnprocessedImageURLs, numWorkers int) (*ProcessedImages, error) {
	b.outputImages = NewProcessedImages()
	b.outputErrs = nil

	queueCh := make(chan UnprocessedImageURL, numWorkers)
	workWg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		go b.worker(&workWg, queueCh)
	}

	for _, unprocessedImageURL := range unprocessedImageURLs.All() {
		workWg.Add(1)
		queueCh <- unprocessedImageURL
	}

	workWg.Wait()
	close(queueCh)

	return b.outputImages, errFromErrs(b.outputErrs)
}

func (b *ImageQueue) worker(workWg *sync.WaitGroup, queueCh <-chan UnprocessedImageURL) {
	for unprocessedImageURL := range queueCh {
		b.work(workWg, unprocessedImageURL)
	}
}

func (b *ImageQueue) work(workWg *sync.WaitGroup, unprocessedImageURL UnprocessedImageURL) {
	defer workWg.Done()

	imgURL, origins, err := b.imgFactory.New(unprocessedImageURL.URL).URL()
	if err != nil {
		b.outputErrsLock.Lock()
		b.outputErrs = append(b.outputErrs, fmt.Errorf("Resolving image '%s': %s", unprocessedImageURL.URL, err))
		b.outputErrsLock.Unlock()
		return
	}

	b.outputImagesLock.Lock()
	b.outputImages.Add(unprocessedImageURL, Image{URL: imgURL, Origins: origins})
	b.outputImagesLock.Unlock()
}
