package handlers

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Calls []CallConfig `yaml:"calls"`
}

type CallConfig struct {
	Method  string            `yaml:"method"`
	Params  map[string]string `yaml:"params"`
	Backend []BackendConfig   `yaml:"backend"`
}

type BackendConfig struct {
	ResponseBody    string   `yaml:"response_body"`
	Allow           []string `yaml:"allow"`
	RequestTemplate string   `yaml:"request_template"`
}

func LoadConfig(path string) (*Config, error) {
	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML data into HandlersConfig
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
