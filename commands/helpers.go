package commands

import (
	"fmt"
	"multirepo/repositories"
	"sort"
)

func PrintRepositoryCounter(config repositories.Config) {
	if len(config.Repos) == 1 {
		fmt.Printf("%d repository detected\n\n", len(config.Repos))
	} else {
		fmt.Printf("%d repositories detected\n\n", len(config.Repos))
	}
}

func GetOrderedRepoNames(config repositories.Config) []string {
	var repoNames []string
	for repoName := range config.Repos {
		repoNames = append(repoNames, repoName)
	}
	sort.Strings(repoNames)
	return repoNames
}
