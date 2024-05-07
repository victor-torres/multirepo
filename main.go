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
	} else if len(os.Args) > 2 && os.Args[1] == "run" {
		err := commands.Run(config, os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Usage: git-subrepos [sync | status | run <command>]")
		os.Exit(1)
	}
}
