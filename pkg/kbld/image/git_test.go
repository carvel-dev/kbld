package image_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	ctlimg "github.com/k14s/kbld/pkg/kbld/image"
)

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
	dir, err := ioutil.TempDir("", "kbld-git-repo")
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
	dir, err := ioutil.TempDir("", "kbld-git-repo")
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

func TestGitRepoValidNonGit(t *testing.T) {
	dir, err := ioutil.TempDir("", "kbld-git-repo")
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

func runCmd(t *testing.T, cmdName string, args []string, dir string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Running cmd %s (%#v): %s", cmdName, args, err)
	}
}
