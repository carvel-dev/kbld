// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	GitRepoRemoteURLUnknown = "<unknown>"
	GitRepoHeadSHANoCommits = "<no commits>"
)

type GitRepo struct {
	dirPath string // CWD when exec-ing to avoid --git-dir weirdness
}

func NewGitRepo(dirPath string) GitRepo {
	return GitRepo{dirPath}
}

func (r GitRepo) RemoteURL() (string, error) {
	// TODO support other remotes besides origin?
	stdout, stderr, err := r.runCmd([]string{"ls-remote", "--get-url"})
	if err != nil {
		// Same message is returned if it's not a git repo
		if r.IsValid() && strings.Contains(stderr, "No remote configured to list refs from") {
			return GitRepoRemoteURLUnknown, nil
		}
		return "", r.error("Determining remote: %s (stderr '%s')", err, stderr)
	}

	return strings.TrimSpace(stdout), nil
}

func (r GitRepo) HeadSHA() (string, error) {
	stdout, stderr, err := r.runCmd([]string{"rev-parse", "HEAD"})
	if err != nil {
		listStdout, _, listErr := r.runCmd([]string{"rev-list", "-n", "1", "--all"})
		if listErr == nil && len(strings.TrimSpace(listStdout)) == 0 {
			return GitRepoHeadSHANoCommits, nil
		}
		return "", r.error("Checking HEAD commit: %s (stderr '%s')", err, stderr)
	}

	return strings.TrimSpace(stdout), nil
}

func (r GitRepo) HeadTags() ([]string, error) {
	stdout, stderr, err := r.runCmd([]string{"describe", "--tags", "--exact-match", "HEAD"})
	if err != nil {
		if strings.Contains(stderr, "no tag exactly matches") ||
			strings.Contains(stderr, "No names found") {
			return nil, nil
		}
		return nil, r.error("Checking HEAD tags: %s (stderr '%s')", err, stderr)
	}

	return strings.Split(strings.TrimSpace(stdout), "\n"), nil
}

func (r GitRepo) IsDirty() (bool, error) {
	stdout, _, err := r.runCmd([]string{"status", "--short"})
	if err != nil {
		return false, r.error("Checking status: %s", err)
	}

	// Strip newline which is added if there are any changes
	return len(strings.TrimSpace(stdout)) > 0, nil
}

func (r GitRepo) IsValid() bool {
	// Prints .git directory path if it's git repo
	_, _, err := r.runCmd([]string{"rev-parse", "--git-dir"})
	return err == nil
}

func (r GitRepo) error(str string, args ...interface{}) error {
	return fmt.Errorf("Getting details from git for directory '%s': %s", r.dirPath, fmt.Sprintf(str, args...))
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

// GitRedactedRemoteURL redacts any user/password information in the URL
func GitRedactedRemoteURL(url string) string {
	pieces := strings.SplitN(url, "@", 2)
	if len(pieces) == 1 {
		return url
	}
	if pieces[0] == "git" {
		return url
	}
	frontPieces := strings.SplitN(pieces[0], "://", 2)
	if len(frontPieces) > 1 {
		return strings.Join(append([]string{frontPieces[0] + "://_redacted_"}, pieces[1:]...), "@")
	}
	return strings.Join(append([]string{"_redacted_"}, pieces[1:]...), "@")
}
