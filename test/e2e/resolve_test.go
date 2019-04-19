package e2e

import (
	"strings"
	"testing"
)

func TestResolveSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: nginx:1.14.2
- image: library/nginx:1.14.2
- image: docker.io/library/nginx:1.14.2
- image: index.docker.io/library/nginx:1.14.2
- image: nginx@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestResolveInvalidDigest(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: nginx@sha256:digest
`

	_, err := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
		AllowError:  true,
	})

	expectedErr := "Expected valid digest reference, but found 'nginx@sha256:digest', reason: digest must be between 71 and 71 runes in length: sha256:digest"

	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", err, expectedErr)
	}
}

func TestResolveUnknownImage(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: unknown
`

	_, err := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
		AllowError:  true,
	})

	expectedErr := "- Resolving image 'unknown': UNAUTHORIZED: authentication required;"

	if !strings.Contains(err.Error(), expectedErr) {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", err, expectedErr)
	}
}

func TestResolveWithOverride(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: unknown
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: unknown
  newImage: docker.io/library/nginx:1.14.2
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
