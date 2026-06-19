package repositories

type Config struct {
	Repos map[string]Repository `yaml:"repositories,flow"`
}

type Repository struct {
	Path   string `yaml:"path"`
	URL    string `yaml:"url"`
	Tag    string `yaml:"tag"`
	Branch string `yaml:"branch"`
	Commit string `yaml:"commit"`
	// Depth limits how much history is cloned (git clone --depth).
	// Zero or absent means a full clone.
	Depth int `yaml:"depth"`
}

type Target struct {
	Type        string
	Name        string
	DisplayName string
}

// LockConfig mirrors repositories.lock: every repository pinned to the
// exact commit it was on when `multirepo lock` ran.
type LockConfig struct {
	Repos map[string]LockedRepository `yaml:"repositories"`
}

type LockedRepository struct {
	Commit string `yaml:"commit"`
}
