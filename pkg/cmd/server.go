package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
)

func runServer(ctx context.Context, cfg *config) error {
	derivAPI := deriv.NewService(&cfg.Deriv)

	connRegistry := repo.NewConnectionRegistry()

	requestHandler, err := core.NewService(&cfg.API, derivAPI, connRegistry)
	if err != nil {
		return fmt.Errorf("failed to create request handler: %w", err)
	}

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
