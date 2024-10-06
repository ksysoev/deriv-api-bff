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
	callhandler, err := core.NewCallHandler(&cfg.API)
	if err != nil {
		return fmt.Errorf("failed to create call handler: %w", err)
	}

	derivAPI := deriv.NewService(&cfg.Deriv)

	connRegistry := repo.NewConnectionRegistry()

	requestHandler := core.NewBackendForFE(derivAPI, callhandler, connRegistry)

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
