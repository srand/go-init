package config

// Representation of a service in the configuration file
type ConfigService struct {
	Name    string         `yaml:"name"`
	Command []string       `yaml:"command"`
	PidFile *ConfigPidFile `yaml:"pidfile"`
}
