package main

func checkoutBase(root, ref, dir string) (cleanup func(), err error) {
	if err := command(root, "git", "fetch", "--quiet", "--no-tags", "--depth=1", "origin", ref).Run(); err != nil {
		return nil, err
	}
	if err := command(root, "git", "worktree", "add", "--quiet", "--detach", dir, "FETCH_HEAD").Run(); err != nil {
		return nil, err
	}
	return func() { command(root, "git", "worktree", "remove", "--force", dir).Run() }, nil
}
