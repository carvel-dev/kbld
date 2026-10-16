// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

// redactURL masks userinfo and query parameters from a URL to prevent
// exposing credentials in error messages. If URL parsing fails, returns
// a safe placeholder.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid URL>"
	}
	if u.User != nil {
		u.User = url.User("_redacted_")
	}
	if u.RawQuery != "" {
		u.RawQuery = "_redacted_"
	}
	return u.String()
}

// redactErrorMessage redacts URLs with embedded credentials from error
// strings. Handles cases where error messages contain redirect URLs or
// other URLs with userinfo (user:password) and query parameters (tokens).
func redactErrorMessage(errMsg string) string {
	const notFound = -1

	// Repeatedly find and redact URLs until none remain
	for {
		// Find https:// or http://
		idx := findScheme(errMsg)
		if idx == notFound {
			break
		}

		// Find URL boundary
		endIdx := findURLBoundary(errMsg, idx)

		// Redact the URL in place
		urlStr := errMsg[idx:endIdx]
		redacted := redactURL(urlStr)
		errMsg = errMsg[:idx] + redacted + errMsg[endIdx:]
	}
	return errMsg
}

// findScheme finds the start of http:// or https:// in a string.
func findScheme(s string) int {
	const notFound = -1

	idx := strings.Index(s, "https://")
	if idx != notFound {
		return idx
	}
	return strings.Index(s, "http://")
}

// findURLBoundary finds where a URL ends (at space, quote, bracket, etc).
func findURLBoundary(s string, start int) int {
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '"', '\'', ' ', ']', ')', ',':
			return i
		}
	}
	return len(s)
}

func (s HTTPFileSource) Description() string {
	return fmt.Sprintf("HTTP URL '%s'", redactURL(s.url))
}

func (s HTTPFileSource) Bytes() ([]byte, error) {
	resp, err := http.Get(s.url)
	if err != nil {
		errMsg := redactErrorMessage(err.Error())
		return nil, fmt.Errorf("Requesting URL '%s': %s",
			redactURL(s.url), errMsg)
	}

	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %s", redactURL(s.url), err)
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
