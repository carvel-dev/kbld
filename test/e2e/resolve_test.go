package e2e

import (
	"io/ioutil"
	"os"
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

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
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

func TestResolveSuccessfulWithAnnotations(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: nginx:1.14.2
- image: library/nginx:1.14.2
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	// TODO dedup same url images
	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - Metas:
        - Tag: 1.14.2
          Type: resolved
          URL: nginx:1.14.2
        URL: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
      - Metas:
        - Tag: 1.14.2
          Type: resolved
          URL: library/nginx:1.14.2
        URL: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
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

	_, err := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
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

	_, err := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
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
- image: final
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: unknown
  newImage: docker.io/library/nginx:1.14.2
- image: final
  newImage: docker.io/library/nginx:1.14.2
  preresolved: true
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: docker.io/library/nginx:1.14.2
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestResolveWithImageMap(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: img1
- image: img2:1.14.2
- image: img3
`

	imageMapData := `{
  "img1": "docker.io/foo/img1@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d",
  "img2:1.14.2": "docker.io/foo/img2:1.14.2",
  "img3": "img3"
}
`

	file, err := ioutil.TempFile("", "kbld-test-resolve-with-image-map")
	if err != nil {
		t.Fatalf("temp file err: %s", err)
	}

	file.Close()
	defer os.RemoveAll(file.Name())

	err = ioutil.WriteFile(file.Name(), []byte(imageMapData), os.ModePerm)
	if err != nil {
		t.Fatalf("write image map err: %s", err)
	}

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "--image-map-file", file.Name()}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
spec:
- image: docker.io/foo/img1@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: docker.io/foo/img2:1.14.2
- image: img3
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
