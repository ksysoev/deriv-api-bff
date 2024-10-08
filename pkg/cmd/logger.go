package cmd

import (
	"log/slog"
	"os"
)

// initLogger initializes the default logger for the application using slog.
// It does not take any parameters.
// It returns an error if the logger initialization fails, although in this implementation, it always returns nil.
func initLogger() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	return nil
}
