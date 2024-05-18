package git

import (
	"fmt"
	"os/exec"
	"strings"

	"multirepo/repositories"
)

func Exists(repo repositories.Repository) bool {
	repoPath, err := repositories.ResolvePath(repo.Path)
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

func IsDirty(repo repositories.Repository) (bool, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
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

func GetCurrentCommit(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "log")
	cmd.Args = append(cmd.Args, "-n")
	cmd.Args = append(cmd.Args, "1")
	cmd.Args = append(cmd.Args, "--pretty=%H")

	out, err := cmd.CombinedOutput()
	outString := string(out)
	outString = strings.TrimSpace(outString)
	return outString, err
}

func GetCurrentBranch(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
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
	repoPath, err := repositories.ResolvePath(repo.Path)
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

func Clone(repo repositories.Repository, recurse bool) error {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clone")
	cmd.Args = append(cmd.Args, repo.URL)
	cmd.Args = append(cmd.Args, repoPath)

	if recurse {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
		cmd.Args = append(cmd.Args, "-j8")
	}

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}

func Stash(repo repositories.Repository) error {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "stash")
	cmd.Args = append(cmd.Args, "-u")

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}

func StashDrop(repo repositories.Repository) error {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "stash")
	cmd.Args = append(cmd.Args, "drop")

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}

func Checkout(repo repositories.Repository, recurse bool) error {
	repoPath, err := repositories.ResolvePath(repo.Path)
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

	if recurse {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
	}

	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", out)
	return err
}
