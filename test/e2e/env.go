package e2e

import (
	"os"
	"strings"
	"testing"
)

type Env struct {
	Namespace       string
	PushDestination string
}

func BuildEnv(t *testing.T) Env {
	env := Env{
		PushDestination: os.Getenv("KBLD_E2E_PUSH_DESTINATION"),
	}
	env.Validate(t)
	return env
}

func (e Env) Validate(t *testing.T) {
	errStrs := []string{}

	if len(e.PushDestination) == 0 {
		errStrs = append(errStrs, "Expected PushDestination to be non-empty")
	}

	if len(errStrs) > 0 {
		t.Fatalf("%s", strings.Join(errStrs, "\n"))
	}
}
