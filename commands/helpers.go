package commands

import (
	"fmt"
	"git-subrepos/repos"
)

func PrintRepositoryCounter(config repos.Config) {
	if len(config.Repos) == 1 {
		fmt.Printf("%d repository detected\n\n", len(config.Repos))
	} else {
		fmt.Printf("%d repositories detected\n\n", len(config.Repos))
	}
}
