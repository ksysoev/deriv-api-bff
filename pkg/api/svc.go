package api

import (
	"context"
	"log/slog"
	"os"

	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/deriv-api-bff/pkg/router"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/server"
)

type Config struct {
	Listen string `mapstructure:"listen"`
}

type Service struct {
	cfg     *Config
	handler wasabi.RequestHandler
}

func NewSevice(cfg *Config, handler wasabi.RequestHandler) *Service {
	return &Service{
		cfg:     cfg,
		handler: handler,
	}
}

func (s *Service) Run(ctx context.Context) error {
	dispatcher := dispatch.NewRouterDispatcher(s.handler, router.Dispatch)
	endpoint := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())
	server := server.NewServer(s.cfg.Listen)
	server.AddChannel(endpoint)

	go func() {
		<-ctx.Done()

		if err := server.Close(); err != nil {
			slog.Error("Fail to close app server", "error", err)
		}
	}()

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}
	return nil
}
