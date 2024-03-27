// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	fileResourcesAllowedExts = []string{".json", ".yaml", ".yml"} // matches kubectl
)

type FileResource struct {
	fileSrc FileSource
}

func NewFileResources(file string) ([]FileResource, error) {
	var fileRs []FileResource

	switch {
	case file == "-":
		fileRs = append(fileRs, FileResource{NewStdinSource()})

	case strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://"):
		fileRs = append(fileRs, FileResource{NewHTTPFileSource(file)})

	default:
		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("Unable to stat file: %s", err)
		}

		if fileInfo.IsDir() {
			var paths []string

			err := filepath.Walk(file, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return err
				}
				ext := filepath.Ext(path)
				for _, allowedExt := range fileResourcesAllowedExts {
					if allowedExt == ext {
						paths = append(paths, path)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("Listing files '%s'", file)
			}

			sort.Strings(paths)

			for _, path := range paths {
				fileRs = append(fileRs, FileResource{NewLocalFileSource(path)})
			}
		} else {
			fileRs = append(fileRs, FileResource{NewLocalFileSource(file)})
		}
	}

	return fileRs, nil
}

func (r FileResource) Description() string { return r.fileSrc.Description() }

func (r FileResource) Resources() ([]Resource, error) {
	docs, err := NewYAMLFile(r.fileSrc).Docs()
	if err != nil {
		return nil, fmt.Errorf("Parsing %s: %s", r.Description(), err)
	}

	var resources []Resource

	for i, doc := range docs {
		rs, err := NewResourcesFromBytes(doc)
		if err != nil {
			return nil, fmt.Errorf("Parsing %s doc %d: %s", r.Description(), i+1, err)
		}

		resources = append(resources, rs...)
	}

	return resources, nil
}
