package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/k14s/kbld/pkg/kbld/version"
)

func TestLockOutputSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

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
  metadata:
  - metas:
    - tag: 1.14.2
      type: resolved
      url: nginx:1.14.2
    url: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  newImage: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
  preresolved: true
- image: sample-app
  metadata:
  - metas:
    - tag: 1.15.1
      type: resolved
      url: nginx:1.15.1
    url: index.docker.io/library/nginx@sha256:4a5573037f358b6cdfa2f3e8a9c33a5cf11bcd1675ca72ca76fbe5bd77d0d682
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
`, "__ver__", version.Version)

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

func TestLockOutputPreservesMetadata(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

	lockFileWithMetas := `
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: 0.22.0
overrides:
- image: nginx
  metadata:
  - metas:
    - Path: /Users/name/workspace/nginx
      Type: local
    - Dirty: true
      RemoteURL: git@github.com:nginx/nginx.git
      SHA: 0284d4f11c7a2a55e03e05784e2d59ee4f02dab0
      Type: git
    url: index.docker.io/library/nginx@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  newImage: index.docker.io/library/nginx@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  preresolved: true
`
	path := "/tmp/kbld-test-lock-output-successful"
	defer os.RemoveAll(path)

	kbld.RunWithOpts([]string{"relocate", "-f", "-", "--repository", env.WithRegistries("docker.io/*username*/kbld-test-relocate"), "--lock-output=" + path}, RunOpts{
		StdinReader: strings.NewReader(lockFileWithMetas),
	})

	// TODO: either a) drop the second override; b) copy metadata to the second override; c) extract metadata and refer to it from within overrides; d) ???
	lockFileWithAppendedMetas := env.WithRegistries(`apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: 0.22.0
overrides:
- image: nginx
  metadata:
  - metas:
    - Path: /Users/name/workspace/nginx
      Type: local
    - Dirty: true
      RemoteURL: git@github.com:nginx/nginx.git
      SHA: 0284d4f11c7a2a55e03e05784e2d59ee4f02dab0
      Type: git
    url: index.docker.io/library/nginx@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  newImage: index.docker.io/*username*/kbld-test-relocate@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  preresolved: true
- image: index.docker.io/library/nginx@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  metadata: []
  newImage: index.docker.io/*username*/kbld-test-relocate@sha256:21f32f6c08406306d822a0e6e8b7dc81f53f336570e852e25fbe1e3e3d0d0133
  preresolved: true
`)

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed while reading %s; err=%s", path, err)
	}

	if string(bs) != lockFileWithAppendedMetas {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", bs, lockFileWithAppendedMetas)
	}
}
