package repositories

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath(path string) (string, error) {
	path = os.Expand(path, func(variable string) string {
		value := os.Getenv(variable)
		if value == "" {
			log.Fatal(fmt.Sprintf("Undefined environment variable '%s'\n", variable))
		}
		return value
	})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	return path, nil
}
