// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lockconfig

import (
	"fmt"
	"io/ioutil"

	regname "github.com/google/go-containerregistry/pkg/name"
	"sigs.k8s.io/yaml"
)

const (
	ImagesLockKind       = "ImagesLock"
	ImagesLockAPIVersion = "imgpkg.carvel.dev/v1alpha1"
)

type ImagesLock struct {
	LockVersion
	Images []ImageRef `json:"images,omitempty"` // This generated yaml, but due to lib we need to use `json`
}

type ImageRef struct {
	Image       string            `json:"image,omitempty"`       // This generated yaml, but due to lib we need to use `json`
	Annotations map[string]string `json:"annotations,omitempty"` // This generated yaml, but due to lib we need to use `json`
}

func NewImagesLockFromPath(path string) (ImagesLock, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return ImagesLock{}, fmt.Errorf("Reading path %s: %s", path, err)
	}

	return NewImagesLockFromBytes(bs)
}

func NewImagesLockFromBytes(data []byte) (ImagesLock, error) {
	var lock ImagesLock

	err := yaml.UnmarshalStrict(data, &lock)
	if err != nil {
		return lock, fmt.Errorf("Unmarshaling images lock: %s", err)
	}

	err = lock.Validate()
	if err != nil {
		return lock, fmt.Errorf("Validating images lock: %s", err)
	}

	return lock, nil
}

func (c ImagesLock) Validate() error {
	if c.APIVersion != ImagesLockAPIVersion {
		return fmt.Errorf("Validating apiVersion: Unknown version (known: %s)", ImagesLockAPIVersion)
	}
	if c.Kind != ImagesLockKind {
		return fmt.Errorf("Validating kind: Unknown kind (known: %s)", ImagesLockKind)
	}
	for _, imageRef := range c.Images {
		if _, err := regname.NewDigest(imageRef.Image); err != nil {
			return fmt.Errorf("Expected ref to be in digest form, got '%s'", imageRef.Image)
		}
	}
	return nil
}

func (c ImagesLock) AsBytes() ([]byte, error) {
	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("Validating images lock: %s", err)
	}

	bs, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("Marshaling config: %s", err)
	}

	return []byte(fmt.Sprintf("---\n%s", bs)), nil
}

func (c ImagesLock) WriteToPath(path string) error {
	bs, err := c.AsBytes()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bs, 0600)
	if err != nil {
		return fmt.Errorf("Writing images config: %s", err)
	}

	return nil
}
