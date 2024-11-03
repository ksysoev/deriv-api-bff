package config

import (
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
)

type BFFService interface {
	UpdateHandlers(handlers map[string]core.Handler)
}

type SourceConfig struct {
	Etcd EtcdConfig `mapstructure:"etcd"`
	Path string     `mapstructure:"path"`
}

type EtcdConfig struct {
	Servers            []string `mapstructure:"servers" yaml:"servers"`
	DialTimeoutSeconds int      `mapstructure:"dialTimeoutSeconds" yaml:"dialTimeoutSeconds"`
}

type ConfigService struct {
	bffService BFFService
	sources    SourceConfig
	curCfg     []handlerfactory.Config
}

func New(cfg SourceConfig, bffService BFFService) (*ConfigService, error) {
	if cfg.Path == "" && cfg.Etcd.Servers == nil {
		return nil, fmt.Errorf("no configuration source provided")
	}

	return &ConfigService{
		sources:    cfg,
		bffService: bffService,
	}, nil
}

func (c *ConfigService) LoadHandlers() error {
	if c.sources.Path != "" {
		fs := NewFileSource(c.sources.Path)
		cfg, err := fs.LoadConfig()

		if err != nil {
			return fmt.Errorf("failed to load config from file: %w", err)
		}

		handlers, err := createHandlers(cfg)
		if err != nil {
			return fmt.Errorf("failed to create handlers: %w", err)
		}

		c.curCfg = cfg

		c.bffService.UpdateHandlers(handlers)
	}

	return nil
}

func createHandlers(cfg []handlerfactory.Config) (map[string]core.Handler, error) {
	handlers := make(map[string]core.Handler, len(cfg))

	for _, c := range cfg {
		name, handler, err := handlerfactory.New(c)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler: %w", err)
		}

		if _, ok := handlers[name]; ok {
			return nil, fmt.Errorf("duplicate handler name: %s", name)
		}

		handlers[name] = handler
	}

	return handlers, nil
}
