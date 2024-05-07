package git

import (
	"fmt"
	"git-subrepos/repos"
	"os/exec"
)

func Exists(repo repos.Repo) bool {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repo.Path)
	cmd.Args = append(cmd.Args, "status")

	_, err := cmd.CombinedOutput()
	return err == nil
}

func Clone(repo repos.Repo) error {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clone")
	cmd.Args = append(cmd.Args, repo.URL)
	cmd.Args = append(cmd.Args, repo.Path)

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}

func Checkout(repo repos.Repo) error {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repo.Path)
	cmd.Args = append(cmd.Args, "checkout")

	if repo.Commit != "" {
		cmd.Args = append(cmd.Args, repo.Commit)
	} else if repo.Tag != "" {
		cmd.Args = append(cmd.Args, repo.Tag)
	} else if repo.Branch != "" {
		cmd.Args = append(cmd.Args, repo.Branch)
	}

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}
