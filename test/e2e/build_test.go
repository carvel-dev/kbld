package e2e

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: docker.io/dkalinin/simple-app-kbld
- image: simple-app-two
- image: simple-app-two-overriden
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/dkalinin/simple-app-kbld
  path: assets/simple-app
- image: simple-app-two
  path: assets/simple-app
# 'unused' should not be built
- image: unused
  path: invalid-dir
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: simple-app-two-overriden
  newImage: simple-app-two
`

	out, _ := kbld.RunWithOpts([]string{"apply", "-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	out = regexp.MustCompile("sha256\\-[a-z0-9]{64}").ReplaceAllString(out, "SHA256-REPLACED")

	expectedOut := `kind: Object
spec:
- image: kbld:docker-io-dkalinin-simple-app-kbld-SHA256-REPLACED
- image: kbld:simple-app-two-SHA256-REPLACED
- image: kbld:simple-app-two-SHA256-REPLACED
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestBuildAndPushSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	input := `
kind: Object
spec:
- image: docker.io/dkalinin/simple-app-kbld
- image: simple-app-two
- image: simple-app-two-overriden
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: docker.io/dkalinin/simple-app-kbld
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
- image: docker.io/dkalinin/simple-app-kbld
- image: simple-app-two
  newImage: docker.io/dkalinin/simple-app-kbld
# 'unused' will not be pushed
- image: unused
  path: invalid-dir
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: simple-app-two-overriden
  newImage: simple-app-two
`

	out, _ := kbld.RunWithOpts([]string{"apply", "-f", "-"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := `kind: Object
spec:
- image: index.docker.io/dkalinin/simple-app-kbld@sha256:47dbbe477076b174fa81418f305a68b6b669cfe94deded82a3000bc6f2d72aa9
- image: index.docker.io/dkalinin/simple-app-kbld@sha256:47dbbe477076b174fa81418f305a68b6b669cfe94deded82a3000bc6f2d72aa9
- image: index.docker.io/dkalinin/simple-app-kbld@sha256:47dbbe477076b174fa81418f305a68b6b669cfe94deded82a3000bc6f2d72aa9
`

	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}
