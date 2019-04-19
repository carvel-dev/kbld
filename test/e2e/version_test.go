package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kapp := Kbld{t, env.Namespace, Logger{}}

	out, _ := kapp.RunWithOpts([]string{"version"}, RunOpts{})

	if !strings.Contains(out, "Client Version") {
		t.Fatalf("Expected to find client version")
	}
}
