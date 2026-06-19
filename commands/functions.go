package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"multirepo/git"
	"multirepo/repositories"

	"github.com/fatih/color"
)

// DefaultJobs caps how many repositories are processed concurrently.
// Git operations are network- and disk-bound rather than CPU-bound, so
// a fixed cap works better than scaling with the core count; 10 stays
// within typical git hosting connection limits.
const DefaultJobs = 10

type repoResult struct {
	output string
	err    error
}

// runPerRepo runs fn for every repository name with at most jobs
// concurrent executions and returns one buffered channel per name, so
// results can be consumed in name order while later repositories are
// still being processed.
func runPerRepo(names []string, jobs int, fn func(string) repoResult) []chan repoResult {
	if jobs < 1 {
		jobs = 1
	}
	semaphore := make(chan struct{}, jobs)
	channels := make([]chan repoResult, len(names))
	for i := range channels {
		channels[i] = make(chan repoResult, 1)
	}
	for i, name := range names {
		go func(i int, name string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			channels[i] <- fn(name)
		}(i, name)
	}
	return channels
}

// syncRepo brings a single repository to its configured target and
// returns the accumulated command output.
func syncRepo(repo repositories.Repository, force bool, recurse bool) (string, error) {
	var buf strings.Builder

	target, err := repositories.ParseTarget(repo)
	if err != nil {
		return buf.String(), err
	}

	exists := git.Exists(repo)
	if !exists {
		// Repository does not exist, let's clone it!
		out, err := git.Clone(repo, recurse)
		buf.WriteString(out)
		if err != nil {
			return buf.String(), err
		}
		// A shallow clone may not contain the target reference
		// (only the default branch tip is fetched), so fall back
		// to a fetch when it is missing.
		if !git.RefExists(repo, target.Name) {
			out, err := git.Fetch(repo)
			buf.WriteString(out)
			if err != nil {
				return buf.String(), err
			}
		}
	} else if target.Type == "branch" || !git.RefExists(repo, target.Name) {
		// Branch targets always fetch so the local branch can be
		// fast-forwarded; tag and commit targets only fetch when the
		// reference is not available locally yet.
		out, err := git.Fetch(repo)
		buf.WriteString(out)
		if err != nil {
			return buf.String(), err
		}
	}

	isDirty, err := git.IsDirty(repo)
	if err != nil {
		return buf.String(), err
	}
	if isDirty {
		if !force {
			return buf.String(), errors.New("uncommitted changes, aborting (use --force to discard them)")
		}
		// Only stash and drop when there is something to stash:
		// on a clean tree the stash creates no entry and the drop
		// would delete a pre-existing stash instead.
		out, err := git.Stash(repo)
		buf.WriteString(out)
		if err != nil {
			return buf.String(), err
		}
		out, err = git.StashDrop(repo)
		buf.WriteString(out)
		if err != nil {
			return buf.String(), err
		}
	}

	out, err := git.Checkout(repo, recurse)
	buf.WriteString(out)
	if err != nil {
		return buf.String(), err
	}

	// A branch checkout lands on the local branch, which may be
	// behind the remote: fast-forward it to origin when possible.
	if target.Type == "branch" && git.RefExists(repo, "origin/"+target.Name) {
		out, err := git.FastForward(repo, target.Name)
		buf.WriteString(out)
		if err != nil {
			return buf.String(), err
		}
	}

	return buf.String(), nil
}

func Sync(config repositories.Config, force bool, recurse bool, jobs int) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)

	results := runPerRepo(orderedRepoNames, jobs, func(repoName string) repoResult {
		output, err := syncRepo(config.Repos[repoName], force, recurse)
		return repoResult{output: output, err: err}
	})

	var errs []error
	for i, repoName := range orderedRepoNames {
		result := <-results[i]
		fmt.Print(result.output)
		fmt.Println()
		if result.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", repoName, result.err))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	return Status(config, "", jobs)
}

// collectRepoStatus gathers the state of a single repository.
func collectRepoStatus(repoName string, repo repositories.Repository) (StatusRow, error) {
	repoPath, err := repositories.ResolvePath(repo.Path)
	if err != nil {
		return StatusRow{}, err
	}

	target, err := repositories.ParseTarget(repo)
	if err != nil {
		return StatusRow{}, err
	}

	row := StatusRow{
		Name:       repoName,
		Path:       repoPath,
		TargetType: target.Type,
		TargetName: target.Name,
	}

	commitHash, err := git.GetCurrentCommit(repo)
	if err != nil {
		return row, nil // Found stays false
	}
	row.Found = true
	row.Commit = commitHash

	row.Dirty, err = git.IsDirty(repo)
	if err != nil {
		return row, fmt.Errorf("failed to check status of repository '%s': %w", repoName, err)
	}

	// Several tags may point at HEAD, one per line
	currentTags, err := git.GetCurrentTags(repo)
	if err != nil {
		return row, fmt.Errorf("failed to list tags of repository '%s': %w", repoName, err)
	}
	if currentTags != "" {
		row.Tags = strings.Split(currentTags, "\n")
	}

	row.Branch, err = git.GetCurrentBranch(repo)
	if err != nil {
		return row, fmt.Errorf("failed to get current branch of repository '%s': %w", repoName, err)
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

	return row, nil
}

// collectStatus gathers the state of every configured repository, at
// most jobs at a time, returning rows in name order.
func collectStatus(config repositories.Config, jobs int) ([]StatusRow, error) {
	names := GetOrderedRepoNames(config)
	rows := make([]StatusRow, len(names))
	collectErrs := make([]error, len(names))

	if jobs < 1 {
		jobs = 1
	}
	semaphore := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i, repoName := range names {
		wg.Add(1)
		go func(i int, repoName string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			rows[i], collectErrs[i] = collectRepoStatus(repoName, config.Repos[repoName])
		}(i, repoName)
	}
	wg.Wait()

	if err := errors.Join(collectErrs...); err != nil {
		return nil, err
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
func Status(config repositories.Config, format string, jobs int) error {
	if format != "" && format != "json" && format != "md" {
		return fmt.Errorf("unknown status format '%s' (supported: json, md)", format)
	}

	rows, err := collectStatus(config, jobs)
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

// Lock resolves the current commit of every repository and writes
// repositories.lock, so a later `sync --locked` reproduces this exact
// state. Every repository must exist locally.
func Lock(config repositories.Config) error {
	PrintRepositoryCounter(config)

	lock := repositories.LockConfig{
		Repos: map[string]repositories.LockedRepository{},
	}
	for _, repoName := range GetOrderedRepoNames(config) {
		repo := config.Repos[repoName]
		commitHash, err := git.GetCurrentCommit(repo)
		if err != nil {
			return fmt.Errorf("cannot lock repository '%s': %w (sync it first)", repoName, err)
		}
		lock.Repos[repoName] = repositories.LockedRepository{Commit: commitHash}
		fmt.Printf("%s ➜ %s\n", repoName, commitHash)
	}

	err := repositories.WriteLock(lock)
	if err != nil {
		return err
	}
	fmt.Printf("\nWrote %s\n", repositories.LockFileName)
	return nil
}
