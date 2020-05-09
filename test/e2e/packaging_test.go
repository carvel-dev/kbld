package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/k14s/kbld/pkg/kbld/version"
)

func TestPkgUnpkgSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	// redis:5.0.4
	input := `
kind: Object
spec:
# references image index
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
# references plain image
- image: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`

	path := "/tmp/kbld-test-pkg-unpkg-successful"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{"unpackage", "-f", "-", "--input", path, "--repository", env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg")}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut = env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`)

	if out != expectedOut {
		t.Fatalf("Expected unpackage output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestPkgUnpkgLockSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
overrides:
# ignored because it's not preresolved
- image: gcs-fetcher
  newImage: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
- image: redis
  newImage: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`

	path := "/tmp/kbld-test-pkg-unpkg-with-lock-successful"
	defer os.RemoveAll(path)

	relocatedLockPath := path + "-relocated"
	defer os.RemoveAll(relocatedLockPath)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{
		"unpackage", "-f", "-", "--input", path, "--repository",
		env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg"),
		"--lock-output", relocatedLockPath,
	}, RunOpts{StdinReader: strings.NewReader(input)})

	lockOutBs, err := ioutil.ReadFile(relocatedLockPath)
	if err != nil {
		t.Fatalf("Expected to find relocated lock file")
	}

	expectedLockOut := strings.ReplaceAll(env.WithRegistries(`apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: __ver__
overrides:
- image: redis
  newImage: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  newImage: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`), "__ver__", version.Version)

	if string(lockOutBs) != expectedLockOut {
		t.Fatalf("Expected unpackage lock output >>>%s<<< to match >>>%s<<<", lockOutBs, expectedLockOut)
	}
}

func TestPkgUnpkgSuccessfulWithForeignLayers(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: index.docker.io/library/mongo@sha256:633ec3ae6db954a65a1abadb482bae73375d0098005cb36a3851b32cd891b22e
`

	path := "/tmp/kbld-test-pkg-unpkg-successful-foreign-layers"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{"unpackage", "-f", "-", "--input", path, "--repository", env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg")}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut = env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:633ec3ae6db954a65a1abadb482bae73375d0098005cb36a3851b32cd891b22e
`)

	if out != expectedOut {
		t.Fatalf("Expected unpackage output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
