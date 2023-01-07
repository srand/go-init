package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// Representation of a configuration file
type ConfigFile struct {
	Services []*ConfigService `yaml:"services"`
	Sysctl   *ConfigSysctl    `yaml:"sysctl"`
	Tasks    []*ConfigTask    `yaml:"tasks"`
}

func (c *ConfigFile) Merge(otherConfig *ConfigFile) error {
	for _, service := range otherConfig.Services {
		c.Services = append(c.Services, service)
	}
	return nil
}

func ParseFile(filepath string) (*ConfigFile, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseReader(f)
}

func ParseReader(reader io.Reader) (*ConfigFile, error) {
	config := ConfigFile{}

	configData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
