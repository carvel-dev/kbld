//go:build e2e

// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

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

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithAnnotations(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	// The repetition in this input is so it can test for:
	// 1) resolving
	// 2) filtering annotations with null origins (which happens with digests, which don't get resolved)
	// 3) de-duplicating annotations
	input := `
kind: Object
spec:
- image: nginx:1.14.2
- image: nginx:1.14.2
- image: library/nginx:1.14.2
- image: docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 1.14.2
            url: nginx:1.14.2
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
      - origins:
        - resolved:
            tag: 1.14.2
            url: library/nginx:1.14.2
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	require.YAMLEq(t, expectedOut, out)
}

func TestSortAnnotations(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: foo/img1:bbb
- image: foo/img1:aaa
- image: foo/img1:ccc
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: foo/img1:bbb
  newImage: bbb
  preresolved: true
- image: foo/img1:aaa
  newImage: aaa
  preresolved: true
- image: foo/img1:ccc
  newImage: ccc
  preresolved: true
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - preresolved:
            url: aaa
        url: aaa
      - origins:
        - preresolved:
            url: bbb
        url: bbb
      - origins:
        - preresolved:
            url: ccc
        url: ccc
spec:
- image: bbb
- image: aaa
- image: ccc
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveInvalidDigest(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx@sha256:digest
`

	_, err := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
		AllowError:  true,
	})

	expectedErr := "Expected valid digest reference, but found 'nginx@sha256:digest', reason: invalid checksum digest length"
	require.Contains(t, err.Error(), expectedErr)
}

func TestResolveUnknownImage(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: unknown
`

	_, err := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
		AllowError:  true,
	})

	expectedErr := "- Resolving image 'unknown': GET https://index.docker.io/v2/library/unknown/manifests/latest: UNAUTHORIZED: authentication required;"
	require.Contains(t, err.Error(), expectedErr)
}

func TestResolveWithOverride(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

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

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveWithImageMap(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

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

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveWithImageKeys(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx:1.14.2
- customImage: nginx:1.14.2
- subPath:
    anotherCustomImage: nginx:1.14.2
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageKeys
keys:
- customImage
- anotherCustomImage
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- customImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- subPath:
    anotherCustomImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveWithOverrideMatchingImageRepo(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: docker.io/library/not-nginx
- image: docker.io/library/not-nginx:1.14.2
- image: docker.io/library/not-nginx:1.20@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- imageRepo: docker.io/library/not-nginx
  newImage: docker.io/library/nginx:1.14.2
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
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithSearchRules(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx:1.14.2
- sidecarImage: index.docker.io/library/nginx:1.14.2
- some_key: nginx:1.14.2
- nestedkey:
    data: |
      nested:
        image: nginx:1.14.2
- images:
  - nginx:1.14.2
  - nginxImage: nginx:1.14.2
    nginxImages:
      value: nginx:1.14.2
- image: skip
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
searchRules:
- keyMatcher:
    name: sidecarImage
- valueMatcher:
    imageRepo: nginx
- keyMatcher:
    name: data
  updateStrategy:
    yaml:
      searchRules:
      - keyMatcher:
          name: image
- keyMatcher:
    path: [spec, {allIndexes: true}, images, {index: 0}]
- keyMatcher:
    path: [spec, {allIndexes: true}, nginxImage]
- keyMatcher:
    path: [spec, {allIndexes: true}, nginxImages, value]
- valueMatcher:
    image: skip
  updateStrategy:
    none: {}
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- sidecarImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- some_key: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- nestedkey:
    data: |
      ---
      nested:
        image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- images:
  - index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  - nginxImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
    nginxImages:
      value: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: skip
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithDuplicateSearchRules(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- sidecarImage: index.docker.io/library/nginx:1.14.2
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
searchRules:
- keyMatcher:
    name: sidecarImage
- keyMatcher:
    name: sidecarImage
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
spec:
- sidecarImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithTagSelection(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx1
- image: nginx2
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: nginx1
  newImage: index.docker.io/library/nginx
  tagSelection:
    semver:
      constraints: "<=1.14.2"
- image: nginx2
  newImage: index.docker.io/library/nginx
  tagSelection:
    semver:
      constraints: "<1.14.2"
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 1.14.1
            url: index.docker.io/library/nginx:1.14.1
        url: index.docker.io/library/nginx@sha256:32fdf92b4e986e109e4db0865758020cb0c3b70d6ba80d02fe87bad5cc3dc228
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:32fdf92b4e986e109e4db0865758020cb0c3b70d6ba80d02fe87bad5cc3dc228
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithPlatformSelectionConfig(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx-arm64
- image: nginx-amd64
- image: nginx-all
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: nginx-arm64
  newImage: index.docker.io/library/nginx:1.14.2
  platformSelection:
    os: linux
    architecture: arm64
- image: nginx-amd64
  newImage: index.docker.io/library/nginx:1.14.2
  platformSelection:
    os: linux
    architecture: amd64
- image: nginx-all
  newImage: index.docker.io/library/nginx:1.14.2
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        - platformSelected:
            architecture: amd64
            index: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
            os: linux
        url: index.docker.io/library/nginx@sha256:706446e9c6667c0880d5da3f39c09a6c7d2114f5a5d6b74a2fafd24ae30d2078
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        - platformSelected:
            architecture: arm64
            index: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
            os: linux
        url: index.docker.io/library/nginx@sha256:d58b3e481b8588c080b42e5d7427f2c2061decbf9194f06e2adce641822e282a
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
spec:
- image: index.docker.io/library/nginx@sha256:d58b3e481b8588c080b42e5d7427f2c2061decbf9194f06e2adce641822e282a
- image: index.docker.io/library/nginx@sha256:706446e9c6667c0880d5da3f39c09a6c7d2114f5a5d6b74a2fafd24ae30d2078
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`

	require.YAMLEq(t, expectedOut, out)
}

func TestResolveSuccessfulWithPlatformSelectionWithGlobalFlag(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: nginx-arm64
- image: nginx-amd64
- image: nginx-all
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: nginx-arm64
  newImage: index.docker.io/library/nginx:1.14.2
  platformSelection:
    os: linux
    architecture: arm64
- image: nginx-amd64
  newImage: index.docker.io/library/nginx:1.14.2
  platformSelection:
    os: linux
    architecture: amd64
# will use flag provided platform selection
- image: nginx-all
  newImage: index.docker.io/library/nginx:1.14.2
`

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--platform", "linux/amd64"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
kind: Object
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        - platformSelected:
            architecture: amd64
            index: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
            os: linux
        url: index.docker.io/library/nginx@sha256:706446e9c6667c0880d5da3f39c09a6c7d2114f5a5d6b74a2fafd24ae30d2078
      - origins:
        - resolved:
            tag: 1.14.2
            url: index.docker.io/library/nginx:1.14.2
        - platformSelected:
            architecture: arm64
            index: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
            os: linux
        url: index.docker.io/library/nginx@sha256:d58b3e481b8588c080b42e5d7427f2c2061decbf9194f06e2adce641822e282a
spec:
- image: index.docker.io/library/nginx@sha256:d58b3e481b8588c080b42e5d7427f2c2061decbf9194f06e2adce641822e282a
- image: index.docker.io/library/nginx@sha256:706446e9c6667c0880d5da3f39c09a6c7d2114f5a5d6b74a2fafd24ae30d2078
- image: index.docker.io/library/nginx@sha256:706446e9c6667c0880d5da3f39c09a6c7d2114f5a5d6b74a2fafd24ae30d2078
`

	require.YAMLEq(t, expectedOut, out)
}
