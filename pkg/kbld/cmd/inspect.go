package cmd

import (
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
}

func NewInspectOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *InspectOptions {
	return &InspectOptions{ui: ui, depsFactory: depsFactory}
}

func NewInspectCmd(o *InspectOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
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
	rs, _, err := o.FileFlags.ResourcesAndConfig()
	if err != nil {
		return err
	}

	foundImages, err := o.findImages(rs)
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
		metasDesc, err := resWithImg.MetasDescription()
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(resWithImg.URL),
			uitable.NewValueString(metasDesc),
			uitable.NewValueString(resWithImg.Resource.Description()),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

func (o *InspectOptions) findImages(rs []ctlres.Resource) ([]foundResourceWithImage, error) {
	foundImages := []foundResourceWithImage{}

	for _, res := range rs {
		visitValues(res.DeepCopyRaw(), imageKey, func(val interface{}) (interface{}, bool) {
			if imgURL, ok := val.(string); ok {
				foundImages = append(foundImages, foundResourceWithImage{URL: imgURL, Resource: res})
			}
			return nil, false
		})
	}

	return foundImages, nil
}

type foundResourceWithImage struct {
	URL      string
	Resource ctlres.Resource
}

func (s foundResourceWithImage) MetasDescription() (string, error) {
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
