package cmd

import (
	"github.com/spf13/cobra"
)

type RegistryFlags struct {
	CACertPaths []string
}

func (s *RegistryFlags) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.CACertPaths, "registry-ca-cert-path", nil, "Add CA certificates for registry API (format: /tmp/foo) (can be specified multiple times)")
}
