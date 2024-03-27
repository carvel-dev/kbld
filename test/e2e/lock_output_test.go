//go:build e2e

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	imgLockWithResolvedOrigins = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: nginx:1.14.2
    kbld.carvel.dev/origins: |
      - resolved:
          tag: 1.14.2
          url: nginx:1.14.2
  image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- annotations:
    kbld.carvel.dev/id: sample-app
    kbld.carvel.dev/origins: |
      - resolved:
          tag: 1.15.1
          url: nginx:1.15.1
  image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
kind: ImagesLock
`
	imgLockWithBuiltOrigins = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: nginx:1.14.2
    kbld.carvel.dev/origins: |
      - local:
          path: path/to/source
      - git:
          dirty: true
          remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
          sha: f7988fb6c02e0ce69257d9bd9cf37ae20a60f1d
  image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- annotations:
    kbld.carvel.dev/id: sample-app
    kbld.carvel.dev/origins: |
      - local:
          path: path/to/source
      - git:
          dirty: true
          remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
          sha: 4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1
  image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
kind: ImagesLock
`
	imgLockWithBuiltAndPreresolvedOrigins = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: nginx:1.14.2
    kbld.carvel.dev/origins: |
      - local:
          path: path/to/source
      - git:
          dirty: true
          remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
          sha: f7988fb6c02e0ce69257d9bd9cf37ae20a60f1d
      - preresolved:
          url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- annotations:
    kbld.carvel.dev/id: sample-app
    kbld.carvel.dev/origins: |
      - local:
          path: path/to/source
      - git:
          dirty: true
          remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
          sha: 4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1
      - preresolved:
          url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
  image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
kind: ImagesLock
`
	imgLock = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
  image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
- annotations:
    kbld.carvel.dev/id: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
kind: ImagesLock
`
)

func TestLockOutputSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
images:
- image: nginx:1.14.2
- image: sample-app
- sidecarImage: sample-app
- - sample-app
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: sample-app
  newImage: nginx:1.15.1
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageKeys
keys:
- sidecarImage
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
searchRules:
- keyMatcher:
    path: [images, {allIndexes: true}, {index: 0}]
`

	path := "/tmp/kbld-test-lock-output-successful"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "--lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	kbldVersionOutput, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})
	kbldVersion := strings.SplitAfter(kbldVersionOutput, " ")[2]

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
- sidecarImage: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
- - index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	expectedFileContents := strings.ReplaceAll(`apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: __ver__
overrides:
- image: nginx:1.14.2
  newImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  preresolved: true
- image: sample-app
  newImage: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
  preresolved: true
searchRules:
- keyMatcher:
    name: sidecarImage
- keyMatcher:
    path:
    - images
    - allIndexes: true
    - index: 0
`, "__ver__", strings.TrimSuffix(kbldVersion, "\n"))

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

func TestImgpkgLockOutputSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
images:
- image: nginx:1.14.2
- image: sample-app
- sidecarImage: sample-app
- - sample-app
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: sample-app
  newImage: nginx:1.15.1
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageKeys
keys:
- sidecarImage
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
searchRules:
- keyMatcher:
    path: [images, {allIndexes: true}, {index: 0}]
`

	path := "/tmp/kbld-test-lock-output-successful"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "--imgpkg-lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
- sidecarImage: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
- - index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed while reading " + path)
	}

	if string(bs) != imgLockWithResolvedOrigins {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", bs, imgLockWithResolvedOrigins)
	}
}

func TestImgpkgLockFileNotInOutput(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := imgLock
	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ``
	if out != expectedOut {
		t.Fatalf("Expected:\n >>>%s<<<\n\nActual:\n >>>%s<<<", expectedOut, out)
	}
}

func TestImgpkgLockFileInputSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
images:
- image: nginx:1.14.2
- image: sample-app
---
` + imgLockWithResolvedOrigins

	out, _ := kbld.RunWithOpts([]string{"-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - resolved:
            tag: 1.15.1
            url: nginx:1.15.1
        - preresolved:
            url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
        url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
      - origins:
        - resolved:
            tag: 1.14.2
            url: nginx:1.14.2
        - preresolved:
            url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestImgpkgLockFileOriginsSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
images:
- image: nginx:1.14.2
- image: sample-app
---
` + imgLockWithBuiltOrigins

	path := "/tmp/kbld-test-lock-origins"
	defer os.RemoveAll(path)
	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--imgpkg-lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
metadata:
  annotations:
    kbld.k14s.io/images: |
      - origins:
        - local:
            path: path/to/source
        - git:
            dirty: true
            remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
            sha: 4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1
        - preresolved:
            url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
        url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
      - origins:
        - local:
            path: path/to/source
        - git:
            dirty: true
            remoteURL: git@github.com:vmware-tanzu/carvel-kbld.git
            sha: f7988fb6c02e0ce69257d9bd9cf37ae20a60f1d
        - preresolved:
            url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
        url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed while reading " + path)
	}

	if string(bs) != imgLockWithBuiltAndPreresolvedOrigins {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", bs, imgLockWithBuiltAndPreresolvedOrigins)
	}
}

func TestImgpkgLockOutputSuccessfulDigestedImageHasNoOrigins(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
`

	path := "/tmp/kbld-test-lock-output-successful"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false", "--imgpkg-lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `---
images:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
- image: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
`
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed while reading " + path)
	}

	// For Digest references, Image Lock should not have origins since there is no image metadata
	if string(bs) != imgLock {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", bs, imgLock)
	}
}
