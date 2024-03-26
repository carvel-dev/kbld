// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type FileSource interface {
	Description() string
	Bytes() ([]byte, error)
}

type StdinSource struct{}

var _ FileSource = StdinSource{}

func NewStdinSource() StdinSource { return StdinSource{} }

func (s StdinSource) Description() string { return "stdin" }

func (s StdinSource) Bytes() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

type LocalFileSource struct {
	path string
}

var _ FileSource = LocalFileSource{}

func NewLocalFileSource(path string) LocalFileSource { return LocalFileSource{path} }

func (s LocalFileSource) Description() string {
	return fmt.Sprintf("file '%s'", s.path)
}

func (s LocalFileSource) Bytes() ([]byte, error) {
	return os.ReadFile(s.path)
}

type HTTPFileSource struct {
	url string
}

var _ FileSource = HTTPFileSource{}

func NewHTTPFileSource(path string) HTTPFileSource { return HTTPFileSource{path} }

func (s HTTPFileSource) Description() string {
	return fmt.Sprintf("HTTP URL '%s'", s.url)
}

func (s HTTPFileSource) Bytes() ([]byte, error) {
	resp, err := http.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, err)
	}

	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %s", s.url, err)
	}

	return result, nil
}

type BytesSource struct {
	bs []byte
}

var _ FileSource = BytesSource{}

func NewBytesSource(bs []byte) BytesSource { return BytesSource{bs} }

func (s BytesSource) Description() string    { return "bytes" }
func (s BytesSource) Bytes() ([]byte, error) { return s.bs, nil }
