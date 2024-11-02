package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func initMetricProvider(ctx context.Context, cfg *config.OtelConfig) error {
	if cfg.Prometheus != nil {
		if err := initPrometheus(ctx, cfg.Prometheus); err != nil {
			return fmt.Errorf("failed to initialize Prometheus: %w", err)
		}
	}

	return nil
}

func initPrometheus(ctx context.Context, cfg *config.PrometheusConfig) error {
	if cfg.Listen == "" || cfg.Path == "" {
		return fmt.Errorf("prometheus listen address and path are required")
	}

	metricExporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
	)

	otel.SetMeterProvider(meterProvider)

	go func() {
		if err := servePrometheus(ctx, cfg); err != nil {
			slog.Error("failed to serve Prometheus", slog.Any("error", err))
		}
	}()

	return nil
}

// servePrometheus serves Prometheus metrics over HTTP.
// It takes a context `ctx` and a Prometheus configuration `cfg` of type *config.PrometheusConfig.
// It returns an error if the server fails to start or if there is an issue closing the server.
// The function listens on the address specified in `cfg.Listen` and serves metrics at the path specified in `cfg.Path`.
// If the context is canceled, the server shuts down gracefully.
func servePrometheus(ctx context.Context, cfg *config.PrometheusConfig) error {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())

	httpSrv := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	slog.Info("serving metrics", slog.Any("listen", cfg.Listen), slog.Any("path", cfg.Path))

	lis, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		if err := httpSrv.Close(); err != nil {
			slog.Error("failed to close metric server", slog.Any("error", err))
		}
	}()

	err = httpSrv.Serve(lis)
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
