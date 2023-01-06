package config

type ConfigTrigger struct {
	Name       string   `yaml:"name"`
	Conditions []string `yaml:"conditions"`
	Actions    []string `yaml:"actions"`
}
