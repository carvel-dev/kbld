package cmd

import (
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type FileFlags struct {
	Files     []string
	Recursive bool
	Sort      bool
}

func (s *FileFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.Files, "file", "f", nil, "Set file (format: /tmp/foo, https://..., -) (can be specified multiple times)")
	cmd.Flags().BoolVarP(&s.Recursive, "recursive", "R", false, "Process directory used in -f recursively")
	cmd.Flags().BoolVar(&s.Sort, "sort", true, "Sort by namespace, name, etc.")
}

func (f *FileFlags) AllResources() ([]ctlres.Resource, error) {
	var rs []ctlres.Resource

	// TODO do anything with kbld configs?
	for _, file := range f.Files {
		fileRs, err := ctlres.NewFileResources(file, f.Recursive)
		if err != nil {
			return nil, err
		}

		for _, fileRes := range fileRs {
			resources, err := fileRes.Resources()
			if err != nil {
				return nil, err
			}

			for _, res := range resources {
				rs = append(rs, res)
			}
		}
	}

	return rs, nil
}

func (f *FileFlags) ResourcesAndConfig() ([]ctlres.Resource, ctlconf.Conf, error) {
	allRs, err := f.AllResources()
	if err != nil {
		return nil, ctlconf.Conf{}, err
	}
	return ctlconf.NewConfFromResources(allRs)
}
