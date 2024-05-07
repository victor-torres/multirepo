package main

import (
	"fmt"
	"log"

	"git-subrepos/git"
	"git-subrepos/repos"
)

func main() {
	config, err := repos.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(config.Repos) == 1 {
		fmt.Printf("%d repository detected\n\n", len(config.Repos))
	} else {
		fmt.Printf("%d repositories detected\n\n", len(config.Repos))
	}

	// Loop through repositories
	for repoName, repo := range config.Repos {
		fmt.Println("Working on", repoName)
		target := repos.ParseTarget(repo)

		// Check if the repository exists
		exists := git.Exists(repo)
		if !exists {
			// Repository does not exist, let's clone it!
			fmt.Println("Repository does not exist at", repo.Path)
			fmt.Printf("Cloning from %s (%s: %s)...\n", repo.URL, target.Type, target.DisplayName)
			err := git.Clone(repo)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("Checking out %s \"%s\"...\n", target.Type, target.DisplayName)
		err = git.Checkout(repo)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println()
	}
}
