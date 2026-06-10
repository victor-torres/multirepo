package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"slices"

	"multirepo/commands"
	"multirepo/repositories"
)

// loadConfig loads environment variables from an optional .env file and
// parses repositories.yaml, exiting on failure. It is only called for
// commands that actually need the config, so usage output stays
// available outside configured directories.
func loadConfig() repositories.Config {
	err := godotenv.Load(".env")
	if err == nil {
		// Stderr so machine-readable output (status --json) stays clean.
		fmt.Fprintf(os.Stderr, "Loading environment variables from .env file\n")
	}

	config, err := repositories.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "sync" {
		force := slices.Contains(os.Args, "--force") || slices.Contains(os.Args, "-f")
		recurse := slices.Contains(os.Args, "--recurse") || slices.Contains(os.Args, "-r")

		err := commands.Sync(loadConfig(), force, recurse)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) >= 2 && os.Args[1] == "status" {
		jsonFlag := slices.Contains(os.Args, "--json")
		mdFlag := slices.Contains(os.Args, "--md")
		if jsonFlag && mdFlag {
			fmt.Println("status: --json and --md are mutually exclusive")
			os.Exit(1)
		}

		var format string
		if jsonFlag {
			format = "json"
		} else if mdFlag {
			format = "md"
		}

		err := commands.Status(loadConfig(), format)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) == 2 && os.Args[1] == "run" {
		fmt.Println("usage: multirepo run <repository name | --all> <command> [<args>]")
		os.Exit(1)
	} else if len(os.Args) > 2 && os.Args[1] == "run" {
		err := commands.Run(loadConfig(), os.Args[2], os.Args[3], os.Args[4:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("usage: multirepo <command> [<args>]")
		fmt.Println()
		fmt.Println("These are common commands used in various situations:")
		fmt.Println()
		fmt.Println("multirepo sync\t\t\t\t\t\t\tClone repositories and checkout the specified reference.")
		fmt.Println("multirepo status [--json | --md]\t\t\t\tDisplay status for each one of the repositories.")
		fmt.Println("multirepo run <repository name | --all> <command> [<args>]\tRun an arbitrary command inside one or all repositories.")
		os.Exit(1)
	}
}
