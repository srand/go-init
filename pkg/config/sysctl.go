package config

type ConfigSysctl struct {
	IncludePaths []string `yaml:"include"`
	Parameters   []struct {
		Key   string
		Value string
	}
}
