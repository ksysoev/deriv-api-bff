package config

import (
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
)

type BFFService interface {
	UpdateHandlers(handlers map[string]core.Handler)
}

type SourceConfig struct {
	Etcd EtcdConfig `mapstructure:"etcd"`
	Path string     `mapstructure:"path"`
}

type ConfigService struct {
	bffService    BFFService
	sources       SourceConfig
	currentConfig []handlerfactory.Config
}

func New(cfg SourceConfig) (*ConfigService, error) {
	if cfg.Path == "" && cfg.Etcd.Servers == nil {
		return nil, fmt.Errorf("no configuration source provided")
	}

	return &ConfigService{
		sources: cfg,
	}, nil
}
