package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"multirepo/repositories"
)

func Exists(repo repositories.Repository) bool {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return false
	}

	// Git walks up parent directories looking for a work tree, so simply
	// running a git command in the path also succeeds for any directory
	// nested inside some other repository. Require the work tree root to
	// be the configured path itself.
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "rev-parse")
	cmd.Args = append(cmd.Args, "--show-toplevel")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	toplevel := strings.TrimSpace(string(out))

	// Compare through EvalSymlinks: git reports the physical path, while
	// the configured path may contain symlinks (e.g. /tmp on macOS).
	resolvedToplevel, err := filepath.EvalSymlinks(toplevel)
	if err != nil {
		return false
	}
	resolvedRepoPath, err := filepath.EvalSymlinks(repoPath)
	if err != nil {
		return false
	}
	return resolvedToplevel == resolvedRepoPath
}

func IsDirty(repo repositories.Repository) (bool, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return false, err
	}

	// --porcelain is stable, locale-independent, and empty exactly when
	// the working tree is clean, unlike the human-readable output.
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "status")
	cmd.Args = append(cmd.Args, "--porcelain")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
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

// RefExists reports whether ref resolves to a commit that is available
// locally. It accepts branches, tags, and (abbreviated) commit hashes.
func RefExists(repo repositories.Repository, ref string) bool {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return false
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "rev-parse")
	cmd.Args = append(cmd.Args, "--verify")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, ref+"^{commit}")

	_, err = cmd.CombinedOutput()
	return err == nil
}

func Fetch(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "fetch")
	cmd.Args = append(cmd.Args, "--tags")

	echo := fmt.Sprintf("➜ %s$ git fetch --tags\n", repoPath)
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}

// FastForward advances the checked-out branch to origin/<branch>. It is
// a no-op when the branch is already up to date or ahead of origin, and
// fails when the histories have diverged (never discards local commits).
func FastForward(repo repositories.Repository, branch string) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "merge")
	cmd.Args = append(cmd.Args, "--ff-only")
	cmd.Args = append(cmd.Args, "origin/"+branch)

	echo := fmt.Sprintf("➜ %s$ git merge --ff-only origin/%s\n", repoPath, branch)
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}

func Clone(repo repositories.Repository, recurse bool) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clone")
	cmd.Args = append(cmd.Args, repo.URL)
	cmd.Args = append(cmd.Args, repoPath)

	if recurse {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
		cmd.Args = append(cmd.Args, "-j8")
	}

	echo := fmt.Sprintf("➜ %s$ git clone %s %s\n", repoPath, repo.URL, strings.Join(cmd.Args[3:], " "))
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}

func Stash(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "stash")
	cmd.Args = append(cmd.Args, "-u")

	echo := fmt.Sprintf("➜ %s$ git stash -u\n", repoPath)
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}

func StashDrop(repo repositories.Repository) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "-C")
	cmd.Args = append(cmd.Args, repoPath)
	cmd.Args = append(cmd.Args, "stash")
	cmd.Args = append(cmd.Args, "drop")

	echo := fmt.Sprintf("➜ %s$ git stash drop\n", repoPath)
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}

func Checkout(repo repositories.Repository, recurse bool) (string, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return "", err
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

	echo := fmt.Sprintf("➜ %s$ git checkout %s\n", repoPath, strings.Join(cmd.Args[4:], " "))
	out, err := cmd.CombinedOutput()
	return echo + string(out), err
}
