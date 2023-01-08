package config

// Representation of a service in the configuration file
type ConfigService struct {
	Name       string              `yaml:"name"`
	CGroup     *ConfigControlGroup `yaml:"cgroup"`
	Command    []string            `yaml:"command"`
	Conditions []string            `yaml:"conditions"`
	PidFile    *ConfigPidFile      `yaml:"pidfile"`
	Triggers   []*ConfigTrigger    `yaml:"triggers"`
}
