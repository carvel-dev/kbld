package cmd

import (
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/ghodss/yaml"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	ui          ui.UI
	depsFactory cmdcore.DepsFactory

	FileFlags     FileFlags
	RegistryFlags RegistryFlags
}

func NewListOptions(ui ui.UI, depsFactory cmdcore.DepsFactory) *ListOptions {
	return &ListOptions{ui: ui, depsFactory: depsFactory}
}

func NewListCmd(o *ListOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List images",
		RunE:    func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	o.FileFlags.Set(cmd)
	o.RegistryFlags.Set(cmd)
	return cmd
}

func (o *ListOptions) Run() error {
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
			uitable.NewHeader("Sources"),
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

	for _, imgSrc := range foundImages {
		srcs, err := imgSrc.Sources()
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []uitable.Value{
			uitable.NewValueString(imgSrc.Image),
			uitable.NewValueString(srcs),
			uitable.NewValueString(imgSrc.Resource.Description()),
		})
	}

	o.ui.PrintTable(table)

	return nil
}

type foundSource struct {
	Image    string
	Resource ctlres.Resource
}

func (s foundSource) Sources() (string, error) {
	metas, err := NewResourceWithBuiltImages(s.Resource.DeepCopyRaw(), nil).Metas()
	if err != nil {
		return "", err
	}

	meta, err := BuiltImageMetas(metas).ForImage(s.Image)
	if err != nil {
		return "", err
	}

	srcsYAML, err := yaml.Marshal(meta.BuiltImageSources)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(srcsYAML)), nil
}

func (o *ListOptions) findImages(rs []ctlres.Resource) ([]foundSource, error) {
	foundImages := []foundSource{}

	for _, res := range rs {
		visitValues(res.DeepCopyRaw(), imageKey, func(val interface{}) (interface{}, bool) {
			if img, ok := val.(string); ok {
				foundImages = append(foundImages, foundSource{Image: img, Resource: res})
			}
			return nil, false
		})
	}

	return foundImages, nil
}
