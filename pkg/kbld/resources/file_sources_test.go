// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"
)

func redactErrorMessageTestCases() []struct {
	name     string
	input    string
	expected string
} {
	return []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "http before https",
			input: `Get "http://u:pw@reg.example/x?t=s": ` +
				`redirect to "https://other.example/y"`,
			expected: `Get "http://_redacted_@reg.example/x?_redacted_": ` +
				`redirect to "https://other.example/y"`,
		},
		{
			name: "https before http",
			input: `Connect to "https://secure.example/api?k=v": ` +
				`fallback "http://bkp:pwd@fallback.example/old"`,
			expected: `Connect to "https://secure.example/api?_redacted_": ` +
				`fallback "http://_redacted_@fallback.example/old"`,
		},
		{
			name: "multiple urls same scheme",
			input: `Failed at "http://u1:pw1@h1.example" and ` +
				`"http://u2:pw2@h2.example"`,
			expected: `Failed at "http://_redacted_@h1.example" and ` +
				`"http://_redacted_@h2.example"`,
		},
		{
			name: "userinfo and query params",
			input: `Request: ` +
				`http://admin:secret@api.example/call?token=abc&session=xyz`,
			expected: `Request: ` +
				`http://_redacted_@api.example/call?_redacted_`,
		},
		{
			name: "url at end with no trailing delimiter",
			input:    `Redirect: http://user:pwd@example.com/resource`,
			expected: `Redirect: http://_redacted_@example.com/resource`,
		},
		{
			name:     "no url in message",
			input:    `Connection timeout after 30 seconds`,
			expected: `Connection timeout after 30 seconds`,
		},
		{
			name: "url with space as delimiter",
			input: `error connecting to ` +
				`http://user:pass@host.example/path next`,
			expected: `error connecting to ` +
				`http://_redacted_@host.example/path next`,
		},
		{
			name: "url in quotes",
			input:    `error: "https://token:value@secure.example/api"`,
			expected: `error: "https://_redacted_@secure.example/api"`,
		},
		{
			name:     "https url with brackets",
			input:    `redirect [https://user:pass@example.com/p] ok`,
			expected: `redirect [https://_redacted_@example.com/p] ok`,
		},
		{
			name: "multiple schemes in same message",
			input: `primary: http://u:p@a.example/x?t=S1 and ` +
				`https://u:p@b.example/y?t=S2`,
			expected: `primary: http://_redacted_@a.example/x?_redacted_ and ` +
				`https://_redacted_@b.example/y?_redacted_`,
		},
	}
}

func TestRedactErrorMessage(t *testing.T) {
	for _, tc := range redactErrorMessageTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			result := redactErrorMessage(tc.input)
			if result != tc.expected {
				t.Errorf("mismatch\ninput:    %s\n"+
					"expected: %s\ngot:      %s",
					tc.input, tc.expected, result)
			}
		})
	}
}

func TestRedactURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "http with userinfo",
			input:    "http://user:password@example.com/path",
			expected: "http://_redacted_@example.com/path",
		},
		{
			name:     "https with query param",
			input:    "https://example.com/api?token=secret123",
			expected: "https://example.com/api?_redacted_",
		},
		{
			name: "both userinfo and query",
			input: "http://admin:secret@api.example/call?" +
				"token=abc&session=xyz",
			expected: "http://_redacted_@api.example/call?_redacted_",
		},
		{
			name:     "no credentials",
			input:    "https://public.example/resource",
			expected: "https://public.example/resource",
		},
		{
			name:     "url without userinfo",
			input:    "https://example.com/secure/path",
			expected: "https://example.com/secure/path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := redactURL(tc.input)
			if result != tc.expected {
				t.Errorf("redactURL mismatch\n"+
					"input:    %s\nexpected: %s\ngot:      %s",
					tc.input, tc.expected, result)
			}
		})
	}
}
