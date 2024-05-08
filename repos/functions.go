package repos

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

func ParseConfig() (Config, error) {
	var config Config

	// Load the YAML file
	yamlFile, err := os.ReadFile("subrepos.yaml")
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

func ParseTarget(repo Repo) (Target, error) {
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
