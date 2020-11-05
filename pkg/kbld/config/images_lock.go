// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
)

const (
	ImagesLockAPIVersion = "imgpkg.carvel.dev/v1alpha1"
	ImagesLockKind       = "ImagesLock"
	ImagesLockKbldID     = "kbld.carvel.dev/id"
)

type ImagesLock struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Spec       ImagesLockSpec `json:"spec"`
}

type ImagesLockSpec struct {
	Images []ImagesLockEntry `json:"images"`
}

type ImagesLockEntry struct {
	Image       string            `json:"image"`
	Annotations map[string]string `json:"annotations"`
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
