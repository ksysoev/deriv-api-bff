package main

import (
	"context"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootCmd := cmd.InitCommands()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("failed to execute command", slog.Any("error", err))
		cancel()
		os.Exit(1)
	}
}
