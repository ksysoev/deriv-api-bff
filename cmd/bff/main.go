package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
)

// version is the version of the application. It should be set at build time.
var version = "dev"
var build = "dev"

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	rootCmd := cmd.InitCommands(build, version)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("failed to execute command", slog.Any("error", err))
		cancel()
		os.Exit(1)
	}

	cancel()
}
