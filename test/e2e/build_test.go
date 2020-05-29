package e2e

import (
	"regexp"
	"strings"
	"testing"
)

func TestDockerBuildSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := env.WithRegistries(`
kind: Object
spec:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: simple-app-two
- image: simple-app-two-overriden
- image: simple-app-three
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/*username*/kbld-e2e-tests-build
  path: assets/simple-app
- image: simple-app-two
  path: assets/simple-app
  docker:
    build:
      file: dev/Dockerfile.dev
- image: simple-app-three
  path: assets/simple-app
  docker:
    build:
      target: build-env
# 'unused' should not be built
- image: unused
  path: invalid-dir
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: simple-app-two-overriden
  newImage: simple-app-two
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	// kbld:192-168-99-100-30777-minikube-tests-kbld-e2e-tests-SHA256-REPLACED
	out = regexp.MustCompile("sha256\\-[a-z0-9]{64}").ReplaceAllString(out, "SHA256-REPLACED")
	out = regexp.MustCompile("kbld:(.+)-kbld-e2e(\\-.*)-SHA256-REPLACED").ReplaceAllString(out, "kbld:img-title-SHA256-REPLACED")

	expectedOut := `---
kind: Object
spec:
- image: kbld:img-title-SHA256-REPLACED
- image: kbld:simple-app-two-SHA256-REPLACED
- image: kbld:simple-app-two-SHA256-REPLACED
- image: kbld:simple-app-three-SHA256-REPLACED
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestDockerBuildAndPushSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := env.WithRegistries(`
kind: Object
spec:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: simple-app-two
- image: simple-app-two-overriden
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/*username*/kbld-e2e-tests-build
  path: assets/simple-app
- image: simple-app-two
  path: assets/simple-app
# 'unused' should not be built
- image: unused
  path: invalid-dir
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: docker.io/*username*/kbld-e2e-tests-build
- image: simple-app-two
  newImage: docker.io/*username*/kbld-e2e-tests-build
# 'unused' will not be pushed
- image: unused
  path: invalid-dir
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: simple-app-two-overriden
  newImage: simple-app-two
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = strings.Replace(out, regexp.MustCompile("sha256:[a-z0-9]{64}").FindString(out), "SHA256-REPLACED", -1)

	expectedOut := env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED
- image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED
- image: index.docker.io/*username*/kbld-e2e-tests-build@SHA256-REPLACED
`)

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

// Should use base image repo for "nice description" within a tag
// to avoid having tags that are too long.
func TestDockerBuildSuccessfulWithImageRepo(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := env.WithRegistries(`
kind: Object
spec:
- image: kbld:simple-app-sha256-402e53420e6919c713b16ed78f09d2024ba25ebf424776072cd253cf044b544f
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- imageRepo: kbld
  path: assets/simple-app
`)

	out, _ := kbld.RunWithOpts([]string{"-f", "-", "--images-annotation=false"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = regexp.MustCompile("sha256\\-[a-z0-9]{64}").ReplaceAllString(out, "SHA256-REPLACED")

	expectedOut := `---
kind: Object
spec:
- image: kbld:kbld-SHA256-REPLACED
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
