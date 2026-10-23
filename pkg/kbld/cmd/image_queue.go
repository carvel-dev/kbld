// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"sync"

	ctlimg "carvel.dev/kbld/pkg/kbld/image"
)

type imageFactory interface {
	New(url string) ctlimg.Image
}

type ImageQueue struct {
	imgFactory imageFactory

	outputImages     *ProcessedImages
	outputImagesLock sync.Mutex

	outputErrs     []error
	outputErrsLock sync.Mutex

	// failFast stops starting new images after the first failure.
	// Images that are already being processed are allowed to finish.
	failFast bool
}

func NewImageQueue(imgFactory imageFactory) *ImageQueue {
	return &ImageQueue{imgFactory: imgFactory}
}

// FailFast makes the queue skip the images that have not been started yet
// as soon as one image fails.
func (b *ImageQueue) FailFast(enabled bool) *ImageQueue {
	b.failFast = enabled
	return b
}

func (b *ImageQueue) hasFailed() bool {
	b.outputErrsLock.Lock()
	defer b.outputErrsLock.Unlock()
	return len(b.outputErrs) > 0
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
		if b.failFast && b.hasFailed() {
			break
		}
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

	if b.failFast && b.hasFailed() {
		return
	}

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
