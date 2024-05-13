package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"multirepo/commands"
	"multirepo/repositories"
)

func main() {
	err := godotenv.Load(".env")
	if err == nil {
		fmt.Printf("Loading environment variables from .env file\n")
	}

	config, err := repositories.ParseConfig()
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
		fmt.Println("usage: multirepo run <repository name | --all> <command> [<args>]")
		os.Exit(1)
	} else if len(os.Args) > 2 && os.Args[1] == "run" {
		err := commands.Run(config, os.Args[2], os.Args[3], os.Args[4:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("usage: multirepo <command> [<args>]")
		fmt.Println()
		fmt.Println("These are common commands used in various situations:")
		fmt.Println()
		fmt.Println("multirepo sync\t\t\t\t\t\t\tClone repositories and checkout the specified reference.")
		fmt.Println("multirepo status\t\t\t\t\t\tDisplay status for each one of the repositories.")
		fmt.Println("multirepo run <repository name | --all> <command> [<args>]\tRun an arbitrary command inside one or all repositories.")
		os.Exit(1)
	}
}
