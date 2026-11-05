// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"sync"
	"testing"

	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlimg "carvel.dev/kbld/pkg/kbld/image"
	"github.com/stretchr/testify/require"
)

const (
	singleWorker = 1
	allImages    = 4
)

type fakeImage struct {
	url   string
	calls *[]string
	lock  *sync.Mutex
}

func (i fakeImage) URL() (string, []ctlconf.Origin, error) {
	i.lock.Lock()
	*i.calls = append(*i.calls, i.url)
	i.lock.Unlock()
	if i.url == "bad" {
		return "", nil, errors.New("build failed")
	}
	return i.url + "@sha256:abc", nil, nil
}

type fakeFactory struct {
	calls []string
	lock  sync.Mutex
}

func (f *fakeFactory) New(url string) ctlimg.Image {
	return fakeImage{url: url, calls: &f.calls, lock: &f.lock}
}

func newURLs() *UnprocessedImageURLs {
	urls := NewUnprocessedImageURLs()
	for _, u := range []string{"bad", "img2", "img3", "img4"} {
		urls.Add(UnprocessedImageURL{u})
	}
	return urls
}

func TestImageQueueFailFast(t *testing.T) {
	factory := &fakeFactory{}
	_, err := NewImageQueue(factory).FailFast(true).Run(newURLs(), singleWorker)
	require.ErrorContains(t, err, "Resolving image 'bad': build failed")
	require.Equal(t, []string{"bad"}, factory.calls)
}

func TestImageQueueWithoutFailFastProcessesEverything(t *testing.T) {
	factory := &fakeFactory{}
	images, err := NewImageQueue(factory).Run(newURLs(), singleWorker)
	require.ErrorContains(t, err, "Resolving image 'bad': build failed")
	require.Len(t, factory.calls, allImages)
	_, found := images.FindByURL(UnprocessedImageURL{"img4"})
	require.True(t, found)
}
