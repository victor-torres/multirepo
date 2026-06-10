package commands

type DirtyStatus struct {
	IsDirty bool
	Reasons []string
	Icon    string
}

// StatusRow is the collected state of one repository, used by every
// status renderer (human, JSON, markdown).
type StatusRow struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Found      bool     `json:"found"`
	Commit     string   `json:"commit,omitempty"`
	Branch     string   `json:"branch,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Dirty      bool     `json:"dirty"`
	TargetType string   `json:"target_type"`
	TargetName string   `json:"target_name"`
	// RefMatches reports whether HEAD satisfies the configured target.
	RefMatches bool `json:"ref_matches"`
	// InSync is RefMatches && !Dirty: nothing to do for this repository.
	InSync bool `json:"in_sync"`
}
