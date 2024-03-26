// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	ctlreg "carvel.dev/kbld/pkg/kbld/registry"
	"github.com/spf13/cobra"
)

type RegistryFlags struct {
	CACertPaths []string
	VerifyCerts bool
	Insecure    bool
}

func (s *RegistryFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.CACertPaths, "registry-ca-cert-path", nil, "Add CA certificates for registry API (format: /tmp/foo) (can be specified multiple times)")
	cmd.Flags().BoolVar(&s.VerifyCerts, "registry-verify-certs", true, "Set whether to verify server's certificate chain and host name")
	cmd.Flags().BoolVar(&s.Insecure, "registry-insecure", false, "Allow the use of http when interacting with registries")
}

func (s *RegistryFlags) AsRegistryOpts() ctlreg.Opts {
	return ctlreg.Opts{
		CACertPaths:   s.CACertPaths,
		VerifyCerts:   s.VerifyCerts,
		Insecure:      s.Insecure,
		EnvAuthPrefix: "KBLD_REGISTRY",
	}
}
