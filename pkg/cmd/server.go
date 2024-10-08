package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
)

// runServer initializes and runs the server with the provided configuration.
// It takes ctx of type context.Context and cfg of type *config.
// It returns an error if the request handler creation fails or if the server fails to run.
func runServer(ctx context.Context, cfg *config) error {
	derivAPI := deriv.NewService(&cfg.Deriv)

	connRegistry := repo.NewConnectionRegistry()

	calls, err := repo.NewCallsRepository(&cfg.API)
	if err != nil {
		return fmt.Errorf("failed to create calls repo: %w", err)
	}

	requestHandler := core.NewService(calls, derivAPI, connRegistry)

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
