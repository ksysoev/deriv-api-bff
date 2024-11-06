package config

import (
	"context"
	"fmt"
	"log/slog"

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
	Watch(ctx context.Context, onUpdate func())
}

type Service struct {
	bff    BFFService
	local  LocalSource
	remote RemoteSource
	curCfg []handlerfactory.Config
}

type Option func(*Service)

// WithLocalSource sets the local source for the service.
// It takes a parameter local of type LocalSource and returns an Option.
// This function modifies the Service instance to use the provided local source.
func WithLocalSource(local LocalSource) Option {
	return func(s *Service) {
		s.local = local
	}
}

// WithRemoteSource sets the remote source for the Service.
// It takes a parameter remote of type RemoteSource and returns an Option.
// This function modifies the Service to use the provided remote source.
func WithRemoteSource(remote RemoteSource) Option {
	return func(s *Service) {
		s.remote = remote
	}
}

// New creates a new Service instance with the provided BFFService and optional configurations.
// It takes a BFFService instance and a variadic number of Option functions to configure the Service.
// It returns a pointer to the created Service and an error if neither local nor remote source is provided.
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

func (c *Service) Start(ctx context.Context) error {
	var (
		cfg []handlerfactory.Config
		err error
	)

	if c.remote != nil {
		cfg, err = c.remote.LoadConfig(ctx)
	} else if c.local != nil {
		cfg, err = c.local.LoadConfig(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := c.processConfig(cfg); err != nil {
		return fmt.Errorf("failed to process config: %w", err)
	}

	if c.remote == nil {
		return nil
	}

	go func() {
		slog.Info("Starting config watcher")
		c.remote.Watch(ctx, func() {
			cfg, err := c.remote.LoadConfig(ctx)
			if err != nil {
				slog.Error("Failed to load handlers from remote source", slog.Any("error", err))
			}

			if err := c.processConfig(cfg); err != nil {
				slog.Error("Failed to process config", slog.Any("error", err))
			}

			slog.Info("Call configuration updated")
		})
	}()

	return nil
}

// LoadHandlers loads the configuration and updates the handlers for the service.
// It takes a context.Context parameter to manage the request lifetime.
// It returns an error if loading the configuration or creating handlers fails.
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

	return c.processConfig(cfg)
}

func (c *Service) processConfig(cfg []handlerfactory.Config) error {
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

// createHandlers initializes a map of handlers based on the provided configuration.
// It takes a slice of handlerfactory.Config as input.
// It returns a map where the keys are handler names (strings) and the values are core.Handler instances.
// It returns an error if a handler cannot be created or if there are duplicate handler names in the configuration.
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
