package commands

import (
	"fmt"
	"sort"
	"strings"

	"multirepo/repositories"

	"github.com/fatih/color"
)

func PrintRepositoryCounter(config repositories.Config) {
	if len(config.Repos) == 1 {
		fmt.Printf("%d repository detected\n\n", len(config.Repos))
	} else {
		fmt.Printf("%d repositories detected\n\n", len(config.Repos))
	}
}

func ParseDirtyStatus(status string, isDirty bool, currentBranch string, currentTags string, target repositories.Target) DirtyStatus {
	var dirtyStatus DirtyStatus
	if isDirty {
		dirtyStatus.IsDirty = true
		dirtyStatus.Reasons = append(dirtyStatus.Reasons, "uncommitted changes")
	}

	if target.Type == "commit" {
		statusQuery := fmt.Sprintf("commit %s", target.Name)
		if !strings.Contains(status, statusQuery) {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reasons = append(dirtyStatus.Reasons, "unmatching commit")
		}
	} else if target.Type == "tag" {
		// FIXME: currentTags might contain more than a single tag string,
		//        we should probably split by \n,
		//        and check if the resulting list contains the expected tag.
		if target.Name != currentTags {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reasons = append(dirtyStatus.Reasons, "unmatching tag")
		}
	} else if target.Type == "branch" {
		if target.Name != currentBranch {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reasons = append(dirtyStatus.Reasons, "unmatching branch")
		}
	}

	if dirtyStatus.IsDirty {
		dirtyStatus.Icon = color.RedString("✗")
	} else {
		dirtyStatus.Icon = color.GreenString("✔")
	}

	return dirtyStatus
}

func GetOrderedRepoNames(config repositories.Config) []string {
	var repoNames []string
	for repoName := range config.Repos {
		repoNames = append(repoNames, repoName)
	}
	sort.Strings(repoNames)
	return repoNames
}
