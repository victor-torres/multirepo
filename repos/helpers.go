package repos

import (
	"os"
	"strings"
	"path/filepath"
)

func ResolveHomeDir(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}

    if strings.HasPrefix(path, "~/") {
        path = filepath.Join(homeDir, path[2:])
    }

    return path, nil
}
