package config

type ConfigModules struct {
	IncludePaths []string `yaml:"include"`
	Modules      []struct {
		Name       string
		Parameters []string
	} `yaml:"probe"`
}
