package config

import (
	"context"
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

type Service struct {
	bffService   BFFService
	localSource  *FileSource
	remoteSource *EtcdSource
	curCfg       []handlerfactory.Config
}

func New(cfg SourceConfig, bffService BFFService) (*Service, error) {
	if cfg.Path == "" && cfg.Etcd.Servers == "" {
		return nil, fmt.Errorf("no configuration source provided")
	}

	svc := &Service{
		bffService: bffService,
	}

	if cfg.Path != "" {
		svc.localSource = NewFileSource(cfg.Path)
	}

	if cfg.Etcd.Servers != "" {
		var err error

		svc.remoteSource, err = NewEtcdSource(cfg.Etcd)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote source: %w", err)
		}
	}

	return svc, nil
}

func (c *Service) LoadHandlers(ctx context.Context) error {
	var (
		cfg []handlerfactory.Config
		err error
	)

	if c.localSource != nil {
		cfg, err = c.localSource.LoadConfig(ctx)
	} else if c.remoteSource != nil {
		cfg, err = c.remoteSource.LoadConfig(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	handlers, err := createHandlers(cfg)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	c.curCfg = cfg

	c.bffService.UpdateHandlers(handlers)

	return nil
}

func (c *Service) PutConfig(ctx context.Context) error {
	if c.remoteSource == nil || c.localSource == nil {
		return fmt.Errorf("local and remote sources are required")
	}

	if c.curCfg == nil {
		if err := c.LoadHandlers(ctx); err != nil {
			return fmt.Errorf("failed to load handlers: %w", err)
		}
	}

	return c.remoteSource.PutConfig(ctx, c.curCfg)
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
