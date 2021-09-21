// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	ctlser "github.com/k14s/kbld/pkg/kbld/search"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ui ui.UI

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
}

func NewInspectOptions(ui ui.UI) *InspectOptions {
	return &InspectOptions{ui: ui}
}

func NewInspectCmd(o *InspectOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect",
		Aliases: []string{"i", "is", "insp"},
		Short:   "Inspect images",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	return cmd
}

func (o *InspectOptions) Run() error {
	rs, conf, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := o.findImages(rs, conf)
	if err != nil {
		return err
	}

	table := uitable.Table{
		Title:   "Images",
		Content: "images",

		Header: []uitable.Header{
			uitable.NewHeader("Image"),
			uitable.NewHeader("Metadata"),
			uitable.NewHeader("Resource"),
		},

		SortBy: []uitable.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 2, Asc: true},
		},

		// Image URLs and other content is too long
		FillFirstColumn: true,
		Transpose:       true,
	}

	for _, resWithImg := range foundImages {
		originsDesc, err := resWithImg.OriginsDescription()
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(resWithImg.URL),
			uitable.NewValueString(originsDesc),
			uitable.NewValueString(resWithImg.Resource.Description()),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

func (o *InspectOptions) findImages(rs []ctlres.Resource,
	conf ctlconf.Conf) ([]foundResourceWithImage, error) {

	foundImages := []foundResourceWithImage{}

	for _, res := range rs {
		imageRefs := ctlser.NewImageRefs(res.DeepCopyRaw(), conf.SearchRules())

		imageRefs.Visit(func(imgURL string) (string, bool) {
			foundImages = append(foundImages, foundResourceWithImage{URL: imgURL, Resource: res})
			return "", false
		})
	}

	return foundImages, nil
}

type foundResourceWithImage struct {
	URL      string
	Resource ctlres.Resource
}

func (s foundResourceWithImage) OriginsDescription() (string, error) {
	images, err := NewResourceWithImages(s.Resource.DeepCopyRaw(), nil).Images()
	if err != nil {
		return "", err
	}

	image, found := Images(images).ForImage(s.URL)
	if !found {
		return "", nil
	}

	return image.Description(), nil
}
