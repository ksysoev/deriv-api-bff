package cmd

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
)

func runServer(ctx context.Context, cfg *config) error {
	callhandler, err := handlers.NewCallHandler(&cfg.API)
	if err != nil {
		return fmt.Errorf("failed to create call handler: %w", err)
	}

	derivAPI := deriv.NewService(&cfg.Deriv)

	requestHandler := handlers.NewBackendForFE(derivAPI, callhandler)

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
