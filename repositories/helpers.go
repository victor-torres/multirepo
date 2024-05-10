package repositories

import (
	"os"
	"path/filepath"
	"strings"
)

func ResolveHomeDir(path string) (string, error) {
	path = os.ExpandEnv(path)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	return path, nil
}
