package image

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepo struct {
	dirPath string // CWD when exec-ing to avoid --git-dir weirdness
}

func NewGitRepo(dirPath string) GitRepo {
	return GitRepo{dirPath}
}

func (r GitRepo) RemoteURL() (string, error) {
	// TODO support other remotes?
	stdout, stderr, err := r.runCmd([]string{"ls-remote", "--get-url"})
	if err != nil {
		if strings.Contains(stderr, "No remote configured to list refs from") {
			return "<unknown>", nil
		}
		return "", fmt.Errorf("Determining remote: %s", err)
	}

	return strings.TrimSpace(stdout), nil
}

func (r GitRepo) HeadSHA() (string, error) {
	stdout, stderr, err := r.runCmd([]string{"rev-parse", "HEAD"})
	if err != nil {
		if strings.Contains(stderr, "Needed a single revision") {
			return "<no commits>", nil
		}
		return "", fmt.Errorf("Checking HEAD commit: %s", err)
	}

	return strings.TrimSpace(stdout), nil
}

func (r GitRepo) HeadTags() ([]string, error) {
	stdout, stderr, err := r.runCmd([]string{"describe", "--tags", "--exact-match", "HEAD"})
	if err != nil {
		if strings.Contains(stderr, "no tag exactly matches") {
			return nil, nil
		}
		return nil, fmt.Errorf("Checking HEAD tags: %s", err)
	}

	return strings.Split(strings.TrimSpace(stdout), "\n"), nil
}

func (r GitRepo) IsDirty() (bool, error) {
	stdout, _, err := r.runCmd([]string{"status", "--short"})
	if err != nil {
		return false, fmt.Errorf("Checking status: %s", err)
	}

	// Strip newline which is added if there are any changes
	return len(strings.TrimSpace(stdout)) > 0, nil
}

func (r GitRepo) IsValid() bool {
	return !r.isNotGitRepo()
}

func (r GitRepo) isNotGitRepo() bool {
	if _, err := os.Stat(filepath.Join(r.dirPath, ".git")); err != nil {
		return false
	}
	_, _, err := r.runCmd([]string{"rev-parse", "--git-dir"})
	return err != nil
}

func (r GitRepo) runCmd(args []string) (string, string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("git", args...)
	cmd.Dir = r.dirPath
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()

	return stdoutBuf.String(), stderrBuf.String(), err
}
