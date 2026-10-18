package main

import (
	"os"
	"os/exec"
	"strings"
)

// repoRoot returns the top-level directory of the git repository containing the working directory.
func repoRoot() (string, error) {
	revParse := exec.Command("git", "rev-parse", "--show-toplevel")
	revParse.Stderr = os.Stderr
	top, err := revParse.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(top)), nil
}

func checkoutBase(root, ref, dir string) (cleanup func(), err error) {
	if err := command(root, "git", "fetch", "--quiet", "--no-tags", "--depth=1", "origin", ref).Run(); err != nil {
		return nil, err
	}
	if err := command(root, "git", "worktree", "add", "--quiet", "--detach", dir, "FETCH_HEAD").Run(); err != nil {
		return nil, err
	}
	return func() { command(root, "git", "worktree", "remove", "--force", dir).Run() }, nil
}
