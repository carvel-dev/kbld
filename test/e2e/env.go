package e2e

import (
	"os"
	"strings"
	"testing"
)

type Env struct {
	Namespace            string
	DockerHubUsername    string
	SkipCFImagesDownload bool
}

func BuildEnv(t *testing.T) Env {
	env := Env{
		DockerHubUsername:    os.Getenv("KBLD_E2E_DOCKERHUB_USERNAME"),
		SkipCFImagesDownload: os.Getenv("KBLD_E2E_SKIP_CF_IMAGES_DOWNLOAD") == "true",
	}
	env.Validate(t)
	return env
}

func (e Env) Validate(t *testing.T) {
	errStrs := []string{}

	if len(e.DockerHubUsername) == 0 {
		errStrs = append(errStrs, "Expected DockerHubUsername to be non-empty")
	}

	if len(errStrs) > 0 {
		t.Fatalf("%s", strings.Join(errStrs, "\n"))
	}
}

func (e Env) WithRegistries(input string) string {
	for _, prefix := range []string{"index.docker.io/", "docker.io/"} {
		input = strings.Replace(input, prefix+"*username*/", prefix+e.DockerHubUsername+"/", -1)
	}
	return input
}
