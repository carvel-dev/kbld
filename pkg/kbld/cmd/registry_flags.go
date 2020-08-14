// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	ctlreg "github.com/k14s/kbld/pkg/kbld/registry"
	"github.com/spf13/cobra"
)

type RegistryFlags struct {
	CACertPaths []string
	VerifyCerts bool
}

func (s *RegistryFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.CACertPaths, "registry-ca-cert-path", nil, "Add CA certificates for registry API (format: /tmp/foo) (can be specified multiple times)")
	cmd.Flags().BoolVar(&s.VerifyCerts, "registry-verify-certs", true, "Set whether to verify server's certificate chain and host name")
}

func (s *RegistryFlags) AsRegistryOpts() ctlreg.RegistryOpts {
	return ctlreg.RegistryOpts{
		CACertPaths:   s.CACertPaths,
		VerifyCerts:   s.VerifyCerts,
		EnvAuthPrefix: "KBLD_REGISTRY",
	}
}
