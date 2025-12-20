package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/config/source"
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

	sourceOpts, err := source.CreateOptions(&cfg.APISource)
	if err != nil {
		return fmt.Errorf("failed to create config source: %w", err)
	}

	cfgSvc, err := config.New(requestHandler, sourceOpts...)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := cfgSvc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start config service: %w", err)
	}

	server, err := api.NewSevice(&cfg.Server, requestHandler)
	if err != nil {
		return err
	}

	return server.Run(ctx)
}
