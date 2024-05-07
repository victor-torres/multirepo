package repos

type Config struct {
	Repos map[string]Repo `yaml:"repos,flow"`
}

type Repo struct {
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
