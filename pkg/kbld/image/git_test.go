// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	ctlimg "carvel.dev/kbld/pkg/kbld/image"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "kbld-git")
	if err != nil {
		log.Fatalf("err creating tmpDir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	gitConfig, err := os.Create(path.Join(tmpDir, ".gitconfig"))
	if err != nil {
		log.Fatalf("err creating gitconfig: %v", err)
	}
	fmt.Fprintf(gitConfig, "[user]\n\t name = kbld\n\temail = foo@example.com")
	gitConfig.Close()
	os.Exit(m.Run())
}

func TestGitRepoValidNonEmptyRepo(t *testing.T) {
	gitRepo := ctlimg.NewGitRepo(".")
	if gitRepo.IsValid() != true {
		t.Fatalf("Expected kbld to be a git repo")
	}
	url, err := gitRepo.RemoteURL()
	if err != nil || len(url) == 0 || url == ctlimg.GitRepoRemoteURLUnknown {
		t.Fatalf("Expected remote url to succeed")
	}
	sha, err := gitRepo.HeadSHA()
	if err != nil || len(sha) == 0 || sha == ctlimg.GitRepoHeadSHANoCommits {
		t.Fatalf("Expected head sha to succeed")
	}
	_, err = gitRepo.HeadTags()
	if err != nil {
		t.Fatalf("Expected head tags to succeed")
	}
	_, err = gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
}

func TestGitRepoValidNoCommit(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)

	gitRepo := ctlimg.NewGitRepo(dir)
	if gitRepo.IsValid() != true {
		t.Fatalf("Expected no commit repo to be a git repo")
	}
	url, err := gitRepo.RemoteURL()
	if err != nil || url != ctlimg.GitRepoRemoteURLUnknown {
		t.Fatalf("Expected remote url")
	}
	sha, err := gitRepo.HeadSHA()
	if err != nil || sha != ctlimg.GitRepoHeadSHANoCommits {
		t.Fatalf("Expected head sha to succeed: %s", err)
	}
	_, err = gitRepo.HeadTags()
	if err != nil {
		t.Fatalf("Expected head tags to succeed")
	}
	_, err = gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
}

func TestGitRepoValidNotOnBranch(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)
	runCmd(t, "git", []string{"commit", "-am", "msg1", "--allow-empty"}, dir)
	runCmd(t, "git", []string{"commit", "-am", "msg2", "--allow-empty"}, dir)
	runCmd(t, "git", []string{"checkout", "HEAD~1"}, dir)

	gitRepo := ctlimg.NewGitRepo(dir)
	if gitRepo.IsValid() != true {
		t.Fatalf("Expected not-on-branch repo to be a git repo")
	}
	url, err := gitRepo.RemoteURL()
	if err != nil || url != ctlimg.GitRepoRemoteURLUnknown {
		t.Fatalf("Expected unknown remote url")
	}
	sha, err := gitRepo.HeadSHA()
	if err != nil || sha == ctlimg.GitRepoHeadSHANoCommits || len(sha) < 20 {
		t.Fatalf("Expected head sha to succeed: %s; %s", err, sha)
	}
	_, err = gitRepo.HeadTags()
	if err != nil {
		t.Fatalf("Expected head tags to succeed")
	}
	_, err = gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
}

func TestGitRepoValidSubDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)
	runCmd(t, "git", []string{"commit", "-am", "msg1", "--allow-empty"}, dir)

	subDir := filepath.Join(dir, "sub-dir")
	err = os.Mkdir(subDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Making subdir: %s", err)
	}

	gitRepo := ctlimg.NewGitRepo(subDir)
	if gitRepo.IsValid() != true {
		t.Fatalf("Expected not-on-branch repo to be a git repo")
	}
	url, err := gitRepo.RemoteURL()
	if err != nil || url != ctlimg.GitRepoRemoteURLUnknown {
		t.Fatalf("Expected unknown remote url")
	}
	sha, err := gitRepo.HeadSHA()
	if err != nil || sha == ctlimg.GitRepoHeadSHANoCommits || len(sha) < 20 {
		t.Fatalf("Expected head sha to succeed: %s; %s", err, sha)
	}
	_, err = gitRepo.HeadTags()
	if err != nil {
		t.Fatalf("Expected head tags to succeed")
	}
	_, err = gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
}

func TestGitRepoValidNonGit(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	gitRepo := ctlimg.NewGitRepo(dir)
	if gitRepo.IsValid() != false {
		t.Fatalf("Expected empty dir to be not a valid git repo")
	}
	_, err = gitRepo.RemoteURL()
	if err == nil {
		t.Fatalf("Did not expect remote url")
	}
	_, err = gitRepo.HeadSHA()
	if err == nil {
		t.Fatalf("Did not expect head sha")
	}
	_, err = gitRepo.HeadTags()
	if err == nil {
		t.Fatalf("Did not expect head tags to succeed")
	}
	_, err = gitRepo.IsDirty()
	if err == nil {
		t.Fatalf("Did not expect dirty to succeed")
	}
}

func TestGitRepoRemoteURL(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)

	gitRepo := ctlimg.NewGitRepo(dir)

	url, err := gitRepo.RemoteURL()
	if err != nil {
		t.Fatalf("Expected url to succeed")
	}
	if url != ctlimg.GitRepoRemoteURLUnknown {
		t.Fatalf("Expected url to be unknown")
	}

	runCmd(t, "git", []string{"remote", "add", "origin", "http://some-remote"}, dir)

	url, err = gitRepo.RemoteURL()
	if err != nil {
		t.Fatalf("Expected url to succeed")
	}
	if url != "http://some-remote" {
		t.Fatalf("Expected url to be correct: %s", url)
	}
}

func TestGitRepoHeadSHA(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)

	gitRepo := ctlimg.NewGitRepo(dir)

	sha, err := gitRepo.HeadSHA()
	if err != nil {
		t.Fatalf("Expected sha to succeed")
	}
	if sha != ctlimg.GitRepoHeadSHANoCommits {
		t.Fatalf("Expected sha to be unknown")
	}

	runCmd(t, "git", []string{"commit", "-am", "msg1", "--allow-empty"}, dir)
	expectedSHAShort := runCmd(t, "git", []string{"log", "-1", "--oneline"}, dir)[0:7]

	sha, err = gitRepo.HeadSHA()
	if err != nil {
		t.Fatalf("Expected sha to succeed")
	}
	if !strings.HasPrefix(sha, expectedSHAShort) {
		t.Fatalf("Expected sha to be correct: %s vs %s", sha, expectedSHAShort)
	}
}

func TestGitRepoIsDirty(t *testing.T) {
	dir, err := os.MkdirTemp("", "kbld-git-repo")
	if err != nil {
		t.Fatalf("Making tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	runCmd(t, "git", []string{"init", "."}, dir)
	runCmd(t, "git", []string{"commit", "-am", "msg1", "--allow-empty"}, dir)

	gitRepo := ctlimg.NewGitRepo(dir)

	dirty, err := gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
	if dirty != false {
		t.Fatalf("Expected dirty to be false")
	}

	runCmd(t, "touch", []string{"file"}, dir)

	dirty, err = gitRepo.IsDirty()
	if err != nil {
		t.Fatalf("Expected dirty to succeed")
	}
	if dirty != true {
		t.Fatalf("Expected dirty to be true")
	}
}

func runCmd(t *testing.T, cmdName string, args []string, dir string) string {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Running cmd %s (%#v): %s (stdout: '%s', stderr: '%s')",
			cmdName, args, err, stdoutBuf.String(), stderrBuf.String())
	}

	return stdoutBuf.String()
}

func TestGitRedactedRemoteURL(t *testing.T) {
	cases := map[string]string{
		"http://something@github.com/my-org/my-repo.git":      "http://_redacted_@github.com/my-org/my-repo.git",
		"https://something@github.com/my-org/my-repo.git":     "https://_redacted_@github.com/my-org/my-repo.git",
		"https://user:password@github.com/my-org/my-repo.git": "https://_redacted_@github.com/my-org/my-repo.git",
		"ssh://user@host.xz/path/to/repo.git/":                "ssh://_redacted_@host.xz/path/to/repo.git/",
		"git@github.com:my-org/my-repo.git":                   "git@github.com:my-org/my-repo.git",
		"github.com:my-org/my-repo.git":                       "github.com:my-org/my-repo.git",
		"user@host.xz/path/to/repo.git/":                      "_redacted_@host.xz/path/to/repo.git/",
	}
	for input, expected := range cases {
		actual := ctlimg.GitRedactedRemoteURL(input)
		if actual != expected {
			t.Fatalf("Expected '%s' to equal '%s' but was '%s'", input, expected, actual)
		}
	}
}
