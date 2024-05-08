package git

import (
	"fmt"
	"os/exec"
	"strings"

	"git-subrepos/repos"
)

func Exists(repo repos.Repo) bool {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repo.Path)
	cmd.Args = append(cmd.Args, "status")

	_, err := cmd.CombinedOutput()
	return err == nil
}

func Status(repo repos.Repo) (string, bool, error) {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repo.Path)
	cmd.Args = append(cmd.Args, "log")
	cmd.Args = append(cmd.Args, "-n")
	cmd.Args = append(cmd.Args, "1")
	cmd.Args = append(cmd.Args, "-b")
	cmd.Args = append(cmd.Args, ".")
	cmd.Args = append(cmd.Args, "--decorate")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	outString = strings.Split(outString, "\n")[0]

	if err != nil {
		return outString, true, err
	}

	cmd = exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repo.Path)
	cmd.Args = append(cmd.Args, "diff")
	cmd.Args = append(cmd.Args, "-q")

	out, err = cmd.CombinedOutput()
	isDirty := len(out) > 1

	return outString, isDirty, err
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
