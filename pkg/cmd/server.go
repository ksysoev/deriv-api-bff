package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/http"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/router"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
)

// runServer initializes and runs the server with the provided configuration.
// It takes ctx of type context.Context and cfg of type *config.
// It returns an error if the request handler creation fails or if the server fails to run.
func runServer(ctx context.Context, cfg *Config) error {
	derivAPI := deriv.NewService(&cfg.Deriv)
	connRegistry := repo.NewConnectionRegistry()
	calls := repo.NewCallsRepository()

	beRouter := router.New(derivAPI, http.NewService())
	requestHandler := core.NewService(calls, beRouter, connRegistry)

	cfgSvc, err := config.New(cfg.APISource, requestHandler)

	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := cfgSvc.LoadHandlers(); err != nil {
		return fmt.Errorf("failed to load handlers: %w", err)
	}

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
