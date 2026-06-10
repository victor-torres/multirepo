package commands

import (
	"encoding/json"
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

		isDirty, err := git.IsDirty(repo)
		if err != nil {
			return err
		}
		if isDirty {
			if !force {
				return fmt.Errorf("repository '%s' has uncommitted changes, aborting (use --force to discard them)", repoName)
			}
			// Only stash and drop when there is something to stash:
			// on a clean tree the stash creates no entry and the drop
			// would delete a pre-existing stash instead.
			if err := git.Stash(repo); err != nil {
				return err
			}
			if err := git.StashDrop(repo); err != nil {
				return err
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

	return Status(config, "")
}

// collectStatus gathers the state of every configured repository in
// name order, without printing anything.
func collectStatus(config repositories.Config) ([]StatusRow, error) {
	var rows []StatusRow
	for _, repoName := range GetOrderedRepoNames(config) {
		repo := config.Repos[repoName]

		repoPath, err := repositories.ResolvePath(repo.Path)
		if err != nil {
			return nil, err
		}

		target, err := repositories.ParseTarget(repo)
		if err != nil {
			return nil, err
		}

		row := StatusRow{
			Name:       repoName,
			Path:       repoPath,
			TargetType: target.Type,
			TargetName: target.Name,
		}

		commitHash, err := git.GetCurrentCommit(repo)
		if err != nil {
			rows = append(rows, row) // Found stays false
			continue
		}
		row.Found = true
		row.Commit = commitHash

		row.Dirty, err = git.IsDirty(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to check status of repository '%s': %w", repoName, err)
		}

		// Several tags may point at HEAD, one per line
		currentTags, err := git.GetCurrentTags(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to list tags of repository '%s': %w", repoName, err)
		}
		if currentTags != "" {
			row.Tags = strings.Split(currentTags, "\n")
		}

		row.Branch, err = git.GetCurrentBranch(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch of repository '%s': %w", repoName, err)
		}

		switch target.Type {
		case "commit":
			// Abbreviated commit hashes are prefixes of the full hash,
			// so prefix matching supports any abbreviation length.
			row.RefMatches = strings.HasPrefix(commitHash, target.Name)
		case "tag":
			row.RefMatches = slices.Contains(row.Tags, target.Name)
		case "branch":
			row.RefMatches = target.Name == row.Branch
		}
		row.InSync = row.RefMatches && !row.Dirty

		rows = append(rows, row)
	}
	return rows, nil
}

// currentReference describes what HEAD currently points at, preferring
// tags, then the branch, then an abbreviated commit hash.
func currentReference(row StatusRow) string {
	if len(row.Tags) > 0 {
		return fmt.Sprintf("tag: %s", strings.Join(row.Tags, ", "))
	}
	if row.Branch != "" {
		return fmt.Sprintf("branch: %s", row.Branch)
	}
	return row.Commit[:7]
}

func renderStatusHuman(rows []StatusRow) {
	maxRepoNameLength := 0
	for _, row := range rows {
		if len(row.Name) > maxRepoNameLength {
			maxRepoNameLength = len(row.Name)
		}
	}

	for _, row := range rows {
		tabString := strings.Repeat(" ", maxRepoNameLength+4-len(row.Name))

		if !row.Found {
			fmt.Printf("%s%s%s\n", row.Name, tabString, color.RedString("✗ repository not found"))
			continue
		}

		icon := color.GreenString("✔")

		var dirtyString string
		if row.Dirty {
			dirtyString = color.RedString("(uncommitted changes)")
			icon = color.RedString("✗")
		}

		var targetString string
		if row.RefMatches {
			if row.TargetType == "tag" || row.TargetType == "branch" {
				targetString = color.GreenString(fmt.Sprintf("(%s: %s)", row.TargetType, row.TargetName))
			}
		} else {
			icon = color.RedString("✗")
			if row.TargetType == "commit" {
				targetString = color.RedString(fmt.Sprintf("(%s ➜ %s)", row.TargetName, row.Commit))
			} else {
				targetString = color.RedString(fmt.Sprintf("(%s: %s ➜ %s)", row.TargetType, row.TargetName, currentReference(row)))
			}
		}

		fmt.Printf("%s%s%s %s %s %s\n", row.Name, tabString, icon, row.Commit, targetString, dirtyString)
	}
}

func renderStatusJSON(rows []StatusRow) error {
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func renderStatusMarkdown(rows []StatusRow) {
	fmt.Println("| repository | ok | commit | target | current | dirty |")
	fmt.Println("| --- | --- | --- | --- | --- | --- |")
	for _, row := range rows {
		if !row.Found {
			fmt.Printf("| %s | ✗ | not found | %s: %s | | |\n", row.Name, row.TargetType, row.TargetName)
			continue
		}
		icon := "✔"
		if !row.InSync {
			icon = "✗"
		}
		dirty := ""
		if row.Dirty {
			dirty = "uncommitted changes"
		}
		fmt.Printf("| %s | %s | %s | %s: %s | %s | %s |\n",
			row.Name, icon, row.Commit, row.TargetType, row.TargetName, currentReference(row), dirty)
	}
}

// Status reports the state of every repository. Supported formats are
// "" (human-readable, colored), "json", and "md" (markdown table).
func Status(config repositories.Config, format string) error {
	if format != "" && format != "json" && format != "md" {
		return fmt.Errorf("unknown status format '%s' (supported: json, md)", format)
	}

	rows, err := collectStatus(config)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return renderStatusJSON(rows)
	case "md":
		renderStatusMarkdown(rows)
	default:
		PrintRepositoryCounter(config)
		renderStatusHuman(rows)
	}
	return nil
}

func Run(config repositories.Config, repository string, command string, args []string) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	var success bool
	var failed []string
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
			// Keep running in the remaining repositories and report
			// every failure at the end.
			failed = append(failed, fmt.Sprintf("%s (%v)", repoName, err))
		}
	}
	if !success {
		return errors.New(fmt.Sprintf("Unknown repository '%s'", repository))
	}
	if len(failed) > 0 {
		return fmt.Errorf("command failed in: %s", strings.Join(failed, ", "))
	}
	return nil
}
