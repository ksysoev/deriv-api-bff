package config

import (
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
)

type Config struct {
	sources []*Source
	Server  api.Config   `mapstructure:"server"`
	Deriv   deriv.Config `mapstructure:"deriv"`
	API     CallsConfig  `mapstructure:"api"`
	Etcd    EtcdConfig   `mapstructure:"etcd"`
	Otel    OtelConfig   `mapstructure:"otel"`
}

type CallsConfig struct {
	Calls []CallConfig `mapstructure:"calls" yaml:"calls"`
}

type EtcdConfig struct {
	Servers            []string `mapstructure:"servers" yaml:"servers"`
	DialTimeoutSeconds int      `mapstructure:"dialTimeoutSeconds" yaml:"dialTimeoutSeconds"`
}

type CallConfig struct {
	Method  string           `mapstructure:"method" yaml:"method"`
	Params  validator.Config `mapstructure:"params" yaml:"params"`
	Backend []*BackendConfig `mapstructure:"backend" yaml:"backend"`
}

type BackendConfig struct {
	Name            string            `mapstructure:"name" yaml:"name"`
	FieldsMap       map[string]string `mapstructure:"fields_map" yaml:"fields_map"`
	ResponseBody    string            `mapstructure:"response_body" yaml:"response_body"`
	RequestTemplate map[string]any    `mapstructure:"request_template" yaml:"request_template"`
	Method          string            `mapstructure:"method" yaml:"method"`
	URLTemplate     string            `mapstructure:"url_template" yaml:"url_template"`
	DependsOn       []string          `mapstructure:"depends_on" yaml:"depends_on"`
	Allow           []string          `mapstructure:"allow" yaml:"allow"`
}

func (cfg *Config) addConfigSource(s Source) {
	if len(cfg.sources) == 0 {
		cfg.sources = make([]*Source, 0)
	}

	cfg.sources = append(cfg.sources, &s)
}

func (cfg *Config) WatchConfig(event *Event[any], key string) error {
	for _, s := range cfg.sources {
		src := *s
		err := src.WatchConfig(event, key)

		if err != nil {
			return err
		}
	}

	return nil
}
