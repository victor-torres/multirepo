package commands

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"multirepo/git"
	"multirepo/repositories"

	"github.com/fatih/color"
)

func Sync(config repositories.Config, force bool, recurse bool) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]
		target, err := repositories.ParseTarget(repo)
		if err != nil {
			return err
		}

		exists := git.Exists(repo)
		if !exists {
			// Repository does not exist, let's clone it!
			err := git.Clone(repo, recurse)
			if err != nil {
				return err
			}
		} else if target.Type == "branch" || !git.RefExists(repo, target.Name) {
			// Branch targets always fetch so the local branch can be
			// fast-forwarded; tag and commit targets only fetch when the
			// reference is not available locally yet.
			err := git.Fetch(repo)
			if err != nil {
				return err
			}
		}

		if force {
			err := git.Stash(repo)
			if err != nil {
				return err
			}
			_ = git.StashDrop(repo)
		} else {
			isDirty, err := git.IsDirty(repo)
			if err != nil {
				return err
			}
			if isDirty {
				return fmt.Errorf("repository '%s' has uncommitted changes, aborting (use --force to discard them)", repoName)
			}
		}

		err = git.Checkout(repo, recurse)
		if err != nil {
			return err
		}

		// A branch checkout lands on the local branch, which may be
		// behind the remote: fast-forward it to origin when possible.
		if target.Type == "branch" && git.RefExists(repo, "origin/"+target.Name) {
			err := git.FastForward(repo, target.Name)
			if err != nil {
				return err
			}
		}

		fmt.Println()
	}

	return Status(config)
}

func Status(config repositories.Config) error {
	PrintRepositoryCounter(config)

	maxRepoNameLength := 0
	for repoName := range config.Repos {
		if len(repoName) > maxRepoNameLength {
			maxRepoNameLength = len(repoName)
		}
	}

	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]

		tabBuilder := &strings.Builder{}
		for i := 0; i < (maxRepoNameLength + 4 - len(repoName)); i++ {
			tabBuilder.WriteString(" ")
		}
		tabString := tabBuilder.String()

		commitHash, err := git.GetCurrentCommit(repo)
		if err != nil {
			fmt.Printf("%s%s%s\n", repoName, tabString, color.RedString("✗ repository not found"))
			continue
		}

		target, err := repositories.ParseTarget(repo)
		if err != nil {
			return err
		}

		icon := color.GreenString("✔")

		var dirtyString string
		isDirty, err := git.IsDirty(repo)
		if isDirty {
			dirtyString = color.RedString("(uncommitted changes)")
			icon = color.RedString("✗")
		}

		// Several tags may point at HEAD, one per line
		currentTags, err := git.GetCurrentTags(repo)
		currentTagList := strings.Split(currentTags, "\n")
		currentBranch, err := git.GetCurrentBranch(repo)

		var currentReference string
		if currentTags != "" {
			currentReference = fmt.Sprintf("tag: %s", strings.Join(currentTagList, ", "))
		} else if currentBranch != "" {
			currentReference = fmt.Sprintf("branch: %s", currentBranch)
		} else {
			currentReference = commitHash[:7]
		}

		var targetString string
		if target.Type == "commit" {
			// Abbreviated commit hashes are prefixes of the full hash,
			// so prefix matching supports any abbreviation length.
			if !strings.HasPrefix(commitHash, target.Name) {
				targetString = color.RedString(fmt.Sprintf("(%s ➜ %s)", target.Name, commitHash))
				icon = color.RedString("✗")
			}
		} else if target.Type == "tag" {
			if slices.Contains(currentTagList, target.Name) {
				targetString = color.GreenString(fmt.Sprintf("(tag: %s)", target.Name))
			} else {
				targetString = color.RedString(fmt.Sprintf("(tag: %s ➜ %s)", target.Name, currentReference))
				icon = color.RedString("✗")
			}
		} else if target.Type == "branch" {
			if target.Name == currentBranch {
				targetString = color.GreenString(fmt.Sprintf("(branch: %s)", currentBranch))
			} else {
				targetString = color.RedString(fmt.Sprintf("(branch: %s ➜ %s)", target.Name, currentReference))
				icon = color.RedString("✗")
			}
		}

		fmt.Printf("%s%s%s %s %s %s\n", repoName, tabString, icon, commitHash, targetString, dirtyString)
	}
	return nil
}

func Run(config repositories.Config, repository string, command string, args []string) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	var success bool
	for _, repoName := range orderedRepoNames {
		if repository != "--all" && repoName != repository {
			continue
		}

		success = true
		repo := config.Repos[repoName]
		repoPath, err := repositories.ResolvePath(repo.Path)
		if err != nil {
			return err
		}
		fmt.Printf("➜ %s$ %s %s\n", repoPath, command, strings.Join(args, " "))

		cmd := exec.Command(command)
		cmd.Dir = repoPath

		for i := 0; i < len(args); i++ {
			cmd.Args = append(cmd.Args, args[i])
		}

		out, err := cmd.CombinedOutput()
		fmt.Println(string(out))
		if err != nil {
			return err
		}
	}
	if !success {
		return errors.New(fmt.Sprintf("Unknown repository '%s'", repository))
	}
	return nil
}
