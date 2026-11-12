// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetar

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeManifestTar(t *testing.T, manifest string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "img.tar")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Creating tar: %s", err)
	}
	defer file.Close()

	writer := tar.NewWriter(file)
	err = writer.WriteHeader(&tar.Header{Name: "manifest.json", Size: int64(len(manifest)), Mode: 0o644})
	if err != nil {
		t.Fatalf("Writing tar header: %s", err)
	}
	if _, err := writer.Write([]byte(manifest)); err != nil {
		t.Fatalf("Writing tar contents: %s", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Closing tar: %s", err)
	}

	return path
}

// manifest.json comes out of the tar, so a descriptor carrying neither an image
// nor an image index is untrusted input. It used to panic with "Unknown item".
func TestReadDescriptorWithNeitherImageNorIndex(t *testing.T) {
	manifests := map[string]string{
		"empty descriptor":       `[{}]`,
		"both fields null":       `[{"Image":null,"ImageIndex":null}]`,
		"unrelated key only":     `[{"Something":1}]`,
		"valid entry then empty": `[{"Image":{"Refs":["example.com/img@sha256:aa"]}},{}]`,
	}

	for name, manifest := range manifests {
		t.Run(name, func(t *testing.T) {
			_, err := NewTarReader(writeManifestTar(t, manifest)).Read()
			if err == nil {
				t.Fatal("Expected an error, got none")
			}
			if !strings.Contains(err.Error(), "have either an image or an image index") {
				t.Errorf("Expected a descriptor error, got: %s", err)
			}
		})
	}
}

func TestReadRejectsInvalidManifestJSON(t *testing.T) {
	_, err := NewTarReader(writeManifestTar(t, `not json`)).Read()
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
}
