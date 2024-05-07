package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Repos map[string]Repo `yaml:"repos,flow"`
}

type Repo struct {
	Path   string `yaml:"path"`
	URL    string `yaml:"url"`
	Tag    string `yaml:"tag"`
	Branch string `yaml:"branch"`
	Commit string `yaml:"commit"`
}

func main() {
	// Setup debug mode
	is_debug := false
	debug, _ := os.LookupEnv("DEBUG")
	debug = strings.ToLower(debug)
	if (debug == "true") || (debug == "t") || (debug == "1") {
		is_debug = true
	}

	// Load the YAML file
	yamlFile, err := os.ReadFile("subrepos.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the config
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	// DEBUG: print the parsed config
	if is_debug {
		fmt.Printf("%#v\n", config)
	}

	// Loop through repositories
	for _, repo := range config.Repos {
		// Print info message
		fmt.Println("Working on", repo.Path)
		// DEBUG: print additional info
		if is_debug {
			fmt.Printf("Branch: %#v\n", repo.Branch)
			fmt.Printf("Tag: %#v\n", repo.Tag)
			fmt.Printf("Commit: %#v\n", repo.Commit)
		}

		// Parse target info
		target_type := ""
		target_name := ""
		if repo.Branch != "" {
			target_type = "branch"
			target_name = repo.Branch
		} else if repo.Tag != "" {
			target_type = "tag"
			target_name = repo.Tag
		} else if repo.Commit != "" {
			target_type = "commit"
			target_name = repo.Commit
		} else {
			target_type = "branch"
			target_name = ""
		}
		target_display_name := target_name
		if target_name == "" {
			target_display_name = "default"
		}

		// Try to open the repository
		var git_repo git.Repository
		r, err := git.PlainOpen(repo.Path)
		git_repo = *r
		if err != nil {
			// Repository does not exist, let's clone it!
			fmt.Println("Repository does not exist at", repo.Path)
			fmt.Printf("Clonning from %s (%s: %s)...\n", repo.URL, target_type, target_display_name)
			r, err := git.PlainClone(repo.Path, false, &git.CloneOptions{
				URL:               repo.URL,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Progress:          os.Stdout,
			})
			git_repo = *r
			if err != nil {
				log.Fatal(err)
			}
		}

		wt, err := git_repo.Worktree()
		if err != nil {
			log.Fatal(err)
		}

		if target_type == "commit" {
			err := wt.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(target_name),
			})
			if err != nil {
				log.Fatal(err)
			}
		} else if target_type == "branch" {
			err := wt.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName(target_name),
			})
			if err != nil {
				log.Fatal(err)
			}
		} else if target_type == "tag" {
			err := wt.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewTagReferenceName(target_name),
			})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Unsupported reference type", target_type)
			log.Fatal()
		}
		fmt.Printf("Checking out %s \"%s\"...\n", target_type, target_display_name)
	}
}
