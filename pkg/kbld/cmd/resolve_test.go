// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	ctlcmd "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/cmd"
	ctlconf "github.com/vmware-tanzu/carvel-kbld/pkg/kbld/config"
)

func TestNewPlatformSelection(t *testing.T) {
	type test struct {
		Input string
		Plat  *ctlconf.PlatformSelection
		Error error
	}

	exs := []test{
		{
			Input: "",
			Error: fmt.Errorf("Parsing platform '': expected format os/arch[/variant]"),
		}, {
			Input: "linux/386",
			Plat: &ctlconf.PlatformSelection{
				Architecture: "386",
				OS:           "linux",
			},
		}, {
			Input: "linux/arm/v7",
			Plat: &ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				Variant:      "v7",
			},
		}, {
			Input: "linux/arm/v6",
			Plat: &ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				Variant:      "v6",
			},
		}, {
			Input: "linux/arm:version",
			Plat: &ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "version",
			},
		}, {
			Input: "linux/arm/v6:version",
			Plat: &ctlconf.PlatformSelection{
				Architecture: "arm",
				OS:           "linux",
				OSVersion:    "version",
				Variant:      "v6",
			},
		}, {
			Input: "linux/arm/v6/another",
			Error: fmt.Errorf("Parsing platform 'linux/arm/v6/another': too many slashes"),
		},
	}

	for _, ex := range exs {
		t.Run(fmt.Sprintf("parsing '%s'", ex.Input), func(t *testing.T) {
			plat, err := ctlcmd.NewPlatformSelection(ex.Input)
			assert.Equal(t, err, ex.Error)
			assert.Equal(t, plat, ex.Plat)
		})
	}
}
