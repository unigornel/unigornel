package config

import "gopkg.in/yaml.v2"

// Config holds the unigornel configuration.
type Config struct {
	GoRoot    string `yaml:"goroot"`
	MiniOS    string `yaml:"minios"`
	Libraries string `yaml:"libraries"`
}

// ParseConfig will parse a Config object from YAML data.
func ParseConfig(data []byte) (Config, error) {
	var c Config
	err := yaml.Unmarshal(data, &c)
	return c, err
}
