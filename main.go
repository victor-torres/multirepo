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

// version is the release version, injected at build time with
// -ldflags "-X main.version=v1.2.3". Local builds report "dev".
var version = "dev"

func printUsage() {
	fmt.Println("usage: multirepo <command> [<args>]")
	fmt.Println()
	fmt.Println("These are common commands used in various situations:")
	fmt.Println()
	fmt.Println("multirepo sync [--locked]\t\t\t\t\tClone repositories and checkout the specified reference.")
	fmt.Println("multirepo lock\t\t\t\t\t\t\tPin every repository to its current commit in repositories.lock.")
	fmt.Println("multirepo status\t\t\t\t\t\tDisplay status for each one of the repositories.")
	fmt.Println("multirepo run <repository name | --all> <command> [<args>]\tRun an arbitrary command inside one or all repositories.")
	fmt.Println("multirepo version\t\t\t\t\t\tPrint the multirepo version.")
	fmt.Println("multirepo help\t\t\t\t\t\t\tShow this message.")
}

// loadConfig loads environment variables from an optional .env file and
// parses repositories.yaml, exiting on failure. It is only called for
// commands that actually need the config, so usage output stays
// available outside configured directories.
func loadConfig() repositories.Config {
	err := godotenv.Load(".env")
	if err == nil {
		fmt.Printf("Loading environment variables from .env file\n")
	}

	config, err := repositories.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("multirepo version %s\n", version)
		return
	} else if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h" || os.Args[1] == "help") {
		printUsage()
		return
	} else if len(os.Args) >= 2 && os.Args[1] == "sync" {
		force := slices.Contains(os.Args, "--force") || slices.Contains(os.Args, "-f")
		recurse := slices.Contains(os.Args, "--recurse") || slices.Contains(os.Args, "-r")
		locked := slices.Contains(os.Args, "--locked")

		config := loadConfig()
		if locked {
			lock, err := repositories.ParseLock()
			if err != nil {
				log.Fatal(err)
			}
			config, err = repositories.ApplyLock(config, lock)
			if err != nil {
				log.Fatal(err)
			}
		}

		err := commands.Sync(config, force, recurse)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) == 2 && os.Args[1] == "lock" {
		err := commands.Lock(loadConfig())
		if err != nil {
			log.Fatal(err)
		}
	} else if len(os.Args) == 2 && os.Args[1] == "status" {
		err := commands.Status(loadConfig())
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
		printUsage()
		os.Exit(1)
	}
}
