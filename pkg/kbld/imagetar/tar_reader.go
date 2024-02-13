// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package imagetar

import (
	"io"

	"github.com/vmware-tanzu/carvel-kbld/pkg/kbld/imagedesc"
)

type TarReader struct {
	path string
}

func NewTarReader(path string) TarReader {
	return TarReader{path}
}

func (r TarReader) Read() ([]imagedesc.ImageOrIndex, error) {
	file := tarFile{r.path}

	manifestFile, err := file.Chunk("manifest.json").Open()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := io.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	ids, err := imagedesc.NewImageRefDescriptorsFromBytes(manifestBytes)
	if err != nil {
		return nil, err
	}

	return imagedesc.NewDescribedReader(ids, file).Read(), nil
}
