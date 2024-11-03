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

type LocalSource interface {
	LoadConfig(ctx context.Context) ([]handlerfactory.Config, error)
}

type RemoteSource interface {
	LoadConfig(ctx context.Context) ([]handlerfactory.Config, error)
	PutConfig(ctx context.Context, cfg []handlerfactory.Config) error
}

type Service struct {
	bff    BFFService
	local  LocalSource
	remote RemoteSource
	curCfg []handlerfactory.Config
}

type Option func(*Service)

func WithLocalSource(local LocalSource) Option {
	return func(s *Service) {
		s.local = local
	}
}

func WithRemoteSource(remote RemoteSource) Option {
	return func(s *Service) {
		s.remote = remote
	}
}

func New(bff BFFService, opts ...Option) (*Service, error) {
	svc := &Service{
		bff: bff,
	}

	for _, opt := range opts {
		opt(svc)
	}

	if svc.local == nil && svc.remote == nil {
		return nil, fmt.Errorf("local or remote source is required")
	}

	return svc, nil
}

func (c *Service) LoadHandlers(ctx context.Context) error {
	var (
		cfg []handlerfactory.Config
		err error
	)

	if c.local != nil {
		cfg, err = c.local.LoadConfig(ctx)
	} else if c.remote != nil {
		cfg, err = c.remote.LoadConfig(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	handlers, err := createHandlers(cfg)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	c.curCfg = cfg

	c.bff.UpdateHandlers(handlers)

	return nil
}

func (c *Service) PutConfig(ctx context.Context) error {
	if c.remote == nil || c.local == nil {
		return fmt.Errorf("local and remote sources are required")
	}

	if c.curCfg == nil {
		if err := c.LoadHandlers(ctx); err != nil {
			return fmt.Errorf("failed to load handlers: %w", err)
		}
	}

	return c.remote.PutConfig(ctx, c.curCfg)
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
