package main

import (
	"fmt"
	"log"
	"os"

	"git-subrepos/commands"
	"git-subrepos/repos"
)

func main() {
	config, err := repos.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 2 && os.Args[1] == "sync" {
		err := commands.Sync(config)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) == 2 && os.Args[1] == "status" {
		err := commands.Status(config)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) == 2 && os.Args[1] == "run" {
		fmt.Println("usage: git-subrepos run <command> [<args>]")
		os.Exit(1)
	} else if len(os.Args) > 2 && os.Args[1] == "run" {
		err := commands.Run(config, os.Args[2], os.Args[3:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("usage: git-subrepos <command> [<args>]")
		fmt.Println()
		fmt.Println("These are common commands used in various situations:")
		fmt.Println()
		fmt.Println("git-subrepos sync\t\t\tClone repositories and checkout the specified revision")
		fmt.Println("git-subrepos status\t\t\tDisplay status for each one of the repositories")
		fmt.Println("git-subrepos run <command> [<args>]\tRun an arbitrary command inside each one of the repositories")
		os.Exit(1)
	}
}
