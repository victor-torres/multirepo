package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath(path string) (string, error) {
	// os.Expand's callback cannot fail, so remember the first undefined
	// variable and report it after expansion.
	var undefinedVariable string
	path = os.Expand(path, func(variable string) string {
		value := os.Getenv(variable)
		if value == "" && undefinedVariable == "" {
			undefinedVariable = variable
		}
		return value
	})
	if undefinedVariable != "" {
		return path, fmt.Errorf("undefined environment variable '%s'", undefinedVariable)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}

	if path == "~" {
		path = homeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	return path, nil
}
