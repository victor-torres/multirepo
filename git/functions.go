package git

import (
	"fmt"
	"os/exec"
	"strings"

	"multirepo/repositories"
)

func Exists(repo repositories.Repository) bool {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return false
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "status")

	_, err = cmd.CombinedOutput()
	return err == nil
}

func Status(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "log")
	cmd.Args = append(cmd.Args, "-n")
	cmd.Args = append(cmd.Args, "1")
	cmd.Args = append(cmd.Args, "-b")
	cmd.Args = append(cmd.Args, ".")
	cmd.Args = append(cmd.Args, "--decorate")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	outString = strings.Split(outString, "\n")[0]

	return outString, err
}

func IsDirty(repo repositories.Repository) (bool, error) {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return false, err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "status")
	cmd.Args = append(cmd.Args, "--long")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	return !strings.Contains(outString, "working tree clean"), err
}

func GetCurrentBranch(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "branch")
	cmd.Args = append(cmd.Args, "--show-current")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	outString = strings.TrimSpace(outString)
	return outString, err
}

func GetCurrentTags(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "tag")
	cmd.Args = append(cmd.Args, "--points-at")
	cmd.Args = append(cmd.Args, "HEAD")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	outString = strings.TrimSpace(outString)
	return outString, err
}

func Clone(repo repositories.Repository) error {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clone")
	cmd.Args = append(cmd.Args, repo.URL)
	cmd.Args = append(cmd.Args, repoPath)

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}

func Checkout(repo repositories.Repository) error {
	repoPath, err := repositories.ResolveHomeDir(repo.Path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
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
