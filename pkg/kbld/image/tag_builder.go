// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

type TagBuilder struct{}

func (d TagBuilder) CheckTagLen128(tag string) string {
	// "A tag ... may contain a maximum of 128 characters."
	// (https://docs.docker.com/engine/reference/commandline/tag/)
	return d.CheckLen(tag, 128)
}

func (d TagBuilder) CheckLen(str string, num int) string {
	if len(str) > num {
		panic(fmt.Sprintf("Expected string '%s' len to be less than %d", str, num))
	}
	return str
}

func (d TagBuilder) TrimStr(str string, num int) string {
	if len(str) > num {
		str = str[:num]
		// Do not end strings on dash
		if strings.HasSuffix(str, "-") {
			str = str[:len(str)-1] + "e"
		}
	}
	return str
}

func (d TagBuilder) RandomStr50() (string, error) {
	bs, err := d.randomBytes(5)
	if err != nil {
		return "", err
	}
	result := ""
	for _, b := range bs {
		result += fmt.Sprintf("%d", b)
	}
	// Timestamp at the beginning for easier sorting
	return d.CheckLen(fmt.Sprintf("rand-%d-%s", time.Now().UTC().UnixNano(), result), 50), nil
}

func (d TagBuilder) randomBytes(n int) ([]byte, error) {
	bs := make([]byte, n)
	_, err := rand.Read(bs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
