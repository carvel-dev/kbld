package e2e

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestLockOutputSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
images:
- image: nginx:1.14.2
- image: sample-app
- sidecarImage: sample-app
`

	inputConfig := `
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: sample-app
  newImage: nginx:1.17.9
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageKeys
keys:
- sidecarImage
`

	path := "/tmp/kbld-test-lock-output-successful"

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "--lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(input + inputConfig),
	})

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:2539d4344dd18e1df02be842ffc435f8e1f699cfc55516e2cf2cb16b7a9aea0b
- sidecarImage: index.docker.io/library/nginx@sha256:2539d4344dd18e1df02be842ffc435f8e1f699cfc55516e2cf2cb16b7a9aea0b
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	expectedFileContents := `apiVersion: kbld.k14s.io/v1alpha1
keys:
- sidecarImage
kind: Config
minimumRequiredVersion: 0.16.0
overrides:
- image: nginx:1.14.2
  newImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  preresolved: true
- image: sample-app
  newImage: index.docker.io/library/nginx@sha256:2539d4344dd18e1df02be842ffc435f8e1f699cfc55516e2cf2cb16b7a9aea0b
  preresolved: true
`

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed while reading " + path)
	}

	if string(bs) != expectedFileContents {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", bs, expectedFileContents)
	}

	out, _ = kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "-f", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
