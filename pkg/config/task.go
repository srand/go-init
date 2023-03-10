package config

// Representation of a task in the configuration file
type ConfigTask struct {
	Name       string              `yaml:"name"`
	CGroup     *ConfigControlGroup `yaml:"cgroup"`
	Command    []string            `yaml:"command"`
	Conditions []string            `yaml:"conditions"`
	Triggers   []*ConfigTrigger    `yaml:"triggers"`
}
