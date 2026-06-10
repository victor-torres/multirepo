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
