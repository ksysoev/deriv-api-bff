package cmd

import (
	"log/slog"
	"os"
)

func initLogger() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	return nil
}
