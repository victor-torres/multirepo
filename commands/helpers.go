package commands

import (
	"fmt"
	"strings"

	"git-subrepos/repos"
)

func PrintRepositoryCounter(config repos.Config) {
	if len(config.Repos) == 1 {
		fmt.Printf("%d repository detected\n\n", len(config.Repos))
	} else {
		fmt.Printf("%d repositories detected\n\n", len(config.Repos))
	}
}

func ParseDirtyStatus(status string, isDirty bool, target repos.Target) DirtyStatus {
	var dirtyStatus DirtyStatus
	dirtyStatus.IsDirty = isDirty

	if target.Type == "commit" {
		statusQuery := fmt.Sprintf("commit %s", target.Name)
		if !strings.Contains(status, statusQuery) {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reason = "unmatching commit"
		}
	} else if target.Type == "tag" {
		statusQuery := fmt.Sprintf("tag: %s", target.Name)
		if !strings.Contains(status, statusQuery) {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reason = "unmatching tag"
		}
	} else if target.Type == "branch" {
		statusQuery := fmt.Sprintf("/%s)", target.Name)
		if !strings.Contains(status, statusQuery) {
			dirtyStatus.IsDirty = true
			dirtyStatus.Reason = "unmatching branch"
		}
	}

	if dirtyStatus.IsDirty && dirtyStatus.Reason == "" {
		dirtyStatus.Reason = "uncommited changes"
	}

	if dirtyStatus.IsDirty {
		dirtyStatus.Icon = "✗"
	} else {
		dirtyStatus.Icon = "✔"
	}

	return dirtyStatus
}
