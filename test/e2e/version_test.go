package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, Logger{}}

	out, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})

	if !strings.Contains(out, "Client Version") {
		t.Fatalf("Expected to find client version")
	}
}
