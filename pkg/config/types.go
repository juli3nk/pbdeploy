package config

type Config struct {
	Description string     `yaml:"description"`
	Repository  Repository `yaml:"repository"`
	Packages    []Package  `yaml:"packages"`
}

type Repository struct {
	Provider string `yaml:"provider"`
	Username string `yaml:"username"`
}

type Package struct {
	Type      string              `yaml:"type"`
	Org       string              `yaml:"org"`
	Name      string              `yaml:"name"`
	Private   bool                `yaml:"private"`
	SrcPath   string              `yaml:"src_path"`
	CreateTag bool                `yaml:"create_tag,omitempty"`
	Options   map[string][]string `yaml:"options,omitempty"`
}
