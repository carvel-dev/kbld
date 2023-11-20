// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
	ctlres "carvel.dev/kbld/pkg/kbld/resources"
	"github.com/spf13/cobra"
)

type FileFlags struct {
	Files     []string
	Recursive bool
	Sort      bool
}

func (s *FileFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.Files, "file", "f", nil, "Set file (format: /tmp/foo, https://..., -) (can be specified multiple times)")
	cmd.Flags().BoolVar(&s.Sort, "sort", true, "Sort by namespace, name, etc.")
}

func (s *FileFlags) AllResources() ([]ctlres.Resource, error) {
	var rs []ctlres.Resource

	// TODO do anything with kbld configs?
	for _, file := range s.Files {
		fileRs, err := ctlres.NewFileResources(file)
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

func (s *FileFlags) ResourcesAndConfig() ([]ctlres.Resource, ctlconf.Conf, error) {
	allRs, err := s.AllResources()
	if err != nil {
		return nil, ctlconf.Conf{}, err
	}
	return ctlconf.NewConfFromResources(allRs)
}
