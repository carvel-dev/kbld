// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	ImagesLockAPIVersion = "imgpkg.carvel.dev/v1alpha1"
	ImagesLockKind       = "ImagesLock"
	ImagesLockKbldID     = "kbld.carvel.dev/id"
)

type ImagesLock struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string
	Spec       ImagesLockSpec
}

type ImagesLockSpec struct {
	Images []ImagesLockEntry
}

type ImagesLockEntry struct {
	Image       string
	Annotations map[string]string
}

func (i ImagesLock) WriteToFile(path string) error {
	bs, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bs, 0600)
	if err != nil {
		return fmt.Errorf("Writing ImagesLock: %s", err)
	}

	return nil
}
