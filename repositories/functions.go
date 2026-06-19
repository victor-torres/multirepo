package repositories

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func ParseConfig() (Config, error) {
	var config Config

	// Load the YAML file
	yamlFile, err := os.ReadFile("repositories.yaml")
	if err != nil {
		return config, err
	}

	// Parse the config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// LockFileName is the lock file written by `multirepo lock` next to
// repositories.yaml.
const LockFileName = "repositories.lock"

// ParseLock reads and parses the repositories.lock file.
func ParseLock() (LockConfig, error) {
	var lock LockConfig

	yamlFile, err := os.ReadFile(LockFileName)
	if err != nil {
		return lock, err
	}

	err = yaml.Unmarshal(yamlFile, &lock)
	if err != nil {
		return lock, err
	}

	return lock, nil
}

// WriteLock serializes the lock and writes it to repositories.lock.
func WriteLock(lock LockConfig) error {
	encoded, err := yaml.Marshal(lock)
	if err != nil {
		return err
	}
	return os.WriteFile(LockFileName, encoded, 0o644)
}

// ApplyLock returns the configuration with every repository's target
// replaced by the exact commit recorded in the lock file. Every
// configured repository must have a lock entry.
func ApplyLock(config Config, lock LockConfig) (Config, error) {
	for repoName, repo := range config.Repos {
		locked, ok := lock.Repos[repoName]
		if !ok || locked.Commit == "" {
			return config, fmt.Errorf("repository '%s' is not in %s (run `multirepo lock` first)", repoName, LockFileName)
		}
		repo.Commit = locked.Commit
		repo.Tag = ""
		repo.Branch = ""
		config.Repos[repoName] = repo
	}
	return config, nil
}

func ParseTarget(repo Repository) (Target, error) {
	var target Target

	if repo.Commit != "" {
		target.Type = "commit"
		target.Name = repo.Commit
		target.DisplayName = repo.Commit
	} else if repo.Tag != "" {
		target.Type = "tag"
		target.Name = repo.Tag
		target.DisplayName = repo.Tag
	} else if repo.Branch != "" {
		target.Type = "branch"
		target.Name = repo.Branch
		target.DisplayName = repo.Branch
	} else {
		return target, errors.New("Missing reference (commit, tag, or branch)")
	}

	return target, nil
}
