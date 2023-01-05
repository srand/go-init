package config

// Representation of a service pidfile configuration
type ConfigPidFile struct {
	// Path to the service's pidfile
	Path string `yaml:"path"`

	// Set true if the pidfile should be created by the init process
	Create bool `yaml:"create"`
}
