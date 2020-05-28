package tarball

import (
	"encoding/json"
	"fmt"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type describedImage struct {
	itd           ImageDescriptor
	layerProvider LayerProvider
}

var _ regv1.Image = describedImage{}

func (i describedImage) Ref() string { return i.itd.Refs[0] }

// Layers returns the ordered collection of filesystem layers that comprise this image.
// The order of the list is oldest/base layer first, and most-recent/top layer last.
func (i describedImage) Layers() ([]regv1.Layer, error) {
	var layers []regv1.Layer
	for _, layerTD := range i.itd.Layers {
		var layer regv1.Layer
		if layerTD.IsDistributable() {
			layerFile, err := i.layerProvider.FindLayer(layerTD)
			if err != nil {
				return nil, err
			}
			layer = describedLayer{layerTD, layerFile}
		} else {
			layer = foreignLayer{layerTD}
		}
		layers = append(layers, layer)
	}
	return layers, nil
}

// MediaType of this image's manifest.
func (i describedImage) MediaType() (types.MediaType, error) {
	return types.MediaType(i.itd.Manifest.MediaType), nil
}

// ConfigName returns the hash of the image's config file.
func (i describedImage) ConfigName() (regv1.Hash, error) {
	return regv1.NewHash(i.itd.Config.Digest)
}

// ConfigFile returns this image's config file.
func (i describedImage) ConfigFile() (*regv1.ConfigFile, error) {
	var config *regv1.ConfigFile
	err := json.Unmarshal([]byte(i.itd.Config.Raw), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// RawConfigFile returns the serialized bytes of ConfigFile()
func (i describedImage) RawConfigFile() ([]byte, error) {
	return []byte(i.itd.Config.Raw), nil
}

// Digest returns the sha256 of this image's manifest.
func (i describedImage) Digest() (regv1.Hash, error) {
	return regv1.NewHash(i.itd.Manifest.Digest)
}

// Manifest returns this image's Manifest object.
func (i describedImage) Manifest() (*regv1.Manifest, error) {
	var manifest *regv1.Manifest
	err := json.Unmarshal([]byte(i.itd.Manifest.Raw), &manifest)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// RawManifest returns the serialized bytes of Manifest()
func (i describedImage) RawManifest() ([]byte, error) {
	return []byte(i.itd.Manifest.Raw), nil
}

func (i describedImage) Size() (int64, error) {
	return int64(len(i.itd.Manifest.Raw)), nil
}

// LayerByDigest returns a Layer for interacting with a particular layer of
// the image, looking it up by "digest" (the compressed hash).
func (i describedImage) LayerByDigest(digest regv1.Hash) (regv1.Layer, error) {
	for _, layerTD := range i.itd.Layers {
		if layerTD.Digest == digest.String() {
			layerFile, err := i.layerProvider.FindLayer(layerTD)
			if err != nil {
				return nil, err
			}
			return describedLayer{layerTD, layerFile}, nil
		}
	}
	return nil, fmt.Errorf("Expected to find layer '%s' by digest", digest)
}

// LayerByDiffID is an analog to LayerByDigest, looking up by "diff id"
// (the uncompressed hash).
func (i describedImage) LayerByDiffID(diffID regv1.Hash) (regv1.Layer, error) {
	for _, layerTD := range i.itd.Layers {
		if layerTD.DiffID == diffID.String() {
			layerFile, err := i.layerProvider.FindLayer(layerTD)
			if err != nil {
				return nil, err
			}
			return describedLayer{layerTD, layerFile}, nil
		}
	}
	return nil, fmt.Errorf("Expected to find layer '%s' by diff id", diffID)
}
