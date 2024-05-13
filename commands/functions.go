package commands

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"multirepo/git"
	"multirepo/repositories"

	"github.com/fatih/color"
)

func Sync(config repositories.Config) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]
		repoPath, err := repositories.ResolvePath(repo.Path)
		if err != nil {
			log.Fatal(err)
		}

		target, err := repositories.ParseTarget(repo)
		if err != nil {
			log.Fatal(err)
		}

		exists := git.Exists(repo)
		if !exists {
			// Repository does not exist, let's clone it!
			fmt.Printf("➜ %s$ git clone %s\n", repoPath, repo.URL)
			err := git.Clone(repo)
			if err != nil {
				return err
			}
		}

		fmt.Printf("➜ %s$ git checkout %s\n", repoPath, target.Name)
		err = git.Checkout(repo)
		if err != nil {
			return err
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

		// FIXME: this might return multiple tags and not just one
		currentTags, err := git.GetCurrentTags(repo)
		currentBranch, err := git.GetCurrentBranch(repo)

		var currentReference string
		if currentTags != "" {
			currentReference = fmt.Sprintf("tag: %s", currentTags)
		} else if currentBranch != "" {
			currentReference = fmt.Sprintf("branch: %s", currentBranch)
		} else {
			currentReference = commitHash
		}

		var targetString string
		if target.Type == "commit" {
			if target.Name != commitHash {
				targetString = color.RedString(fmt.Sprintf("(%s ➜ %s)", target.Name, commitHash))
				icon = color.RedString("✗")
			}
		} else if target.Type == "tag" {
			if target.Name == currentTags {
				targetString = color.GreenString(fmt.Sprintf("(tag: %s)", currentTags))
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

func Run(config repositories.Config, command string, args []string) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]
		repoPath, err := repositories.ResolvePath(repo.Path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("➜ %s$ %s %s\n", repoPath, command, strings.Join(args, " "))

		cmd := exec.Command(command)
		cmd.Dir = repoPath

		for i := 0; i < len(args); i++ {
			cmd.Args = append(cmd.Args, args[i])
		}

		out, err := cmd.CombinedOutput()
		outString := string(out)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(outString)
	}
	return nil
}
