package commands

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"git-subrepos/git"
	"git-subrepos/repos"

	"github.com/fatih/color"
)

func Sync(config repos.Config) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]
		target, err := repos.ParseTarget(repo)
		if err != nil {
			log.Fatal(err)
		}

		exists := git.Exists(repo)
		if !exists {
			// Repository does not exist, let's clone it!
			fmt.Printf("➜ %s$ git clone %s\n", repo.Path, repo.URL)
			err := git.Clone(repo)
			if err != nil {
				return err
			}
		}

		fmt.Printf("➜ %s$ git checkout %s\n", repo.Path, target.Name)
		err = git.Checkout(repo)
		if err != nil {
			return err
		}

		fmt.Println()
	}

	return Status(config)
}

func Status(config repos.Config) error {
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

		status, err := git.Status(repo)
		if err != nil {
			fmt.Printf("%s%s%s\n", repoName, tabString, color.RedString("✗ repository not found"))
			continue
		}

		target, err := repos.ParseTarget(repo)
		if err != nil {
			return err
		}

		isDirty, err := git.IsDirty(repo)
		if err != nil {
			return err
		}
		dirtyStatus := ParseDirtyStatus(status, isDirty, target)
		reasons := strings.Join(dirtyStatus.Reasons, ", ")
		fmt.Printf("%s%s%s %s %s\n", repoName, tabString, dirtyStatus.Icon, status, color.RedString(reasons))
	}
	return nil
}

func Run(config repos.Config, command string, args []string) error {
	PrintRepositoryCounter(config)
	orderedRepoNames := GetOrderedRepoNames(config)
	for _, repoName := range orderedRepoNames {
		repo := config.Repos[repoName]
		fmt.Printf("➜ %s$ %s %s\n", repo.Path, command, strings.Join(args, " "))

		cmd := exec.Command(command)
		cmd.Dir = repo.Path

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
