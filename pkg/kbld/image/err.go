// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	ctlconf "carvel.dev/kbld/pkg/kbld/config"
)

type ErrImage struct {
	err error
}

var _ Image = ErrImage{}

func NewErrImage(err error) ErrImage { return ErrImage{err} }

func (i ErrImage) URL() (string, []ctlconf.Origin, error) { return "", nil, i.err }
