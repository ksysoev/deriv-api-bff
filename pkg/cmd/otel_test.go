package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestServePrometheus(t *testing.T) {
	tests := []struct {
		cfg     *config.PrometheusConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.PrometheusConfig{
				Listen: ":0",
				Path:   "/metrics",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty listen",
			cfg: &config.PrometheusConfig{
				Listen: ":InvalidPort",
				Path:   "/metrics",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			err := servePrometheus(ctx, tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestInitPrometheus(t *testing.T) {
	tests := []struct {
		cfg     *config.PrometheusConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.PrometheusConfig{
				Listen: ":0",
				Path:   "/metrics",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty listen",
			cfg: &config.PrometheusConfig{
				Listen: "",
				Path:   "/metrics",
			},
			wantErr: true,
		},
		{
			name: "invalid config - empty path",
			cfg: &config.PrometheusConfig{
				Listen: ":0",
				Path:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid config - empty listen and path",
			cfg: &config.PrometheusConfig{
				Listen: "",
				Path:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := initPrometheus(ctx, tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestInitMetricProvider(t *testing.T) {
	tests := []struct {
		cfg     *config.OtelConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config with Prometheus",
			cfg: &config.OtelConfig{
				Prometheus: &config.PrometheusConfig{
					Listen: ":0",
					Path:   "/metrics",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid Prometheus config - empty listen",
			cfg: &config.OtelConfig{
				Prometheus: &config.PrometheusConfig{
					Listen: "",
					Path:   "/metrics",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid Prometheus config - empty path",
			cfg: &config.OtelConfig{
				Prometheus: &config.PrometheusConfig{
					Listen: ":0",
					Path:   "",
				},
			},
			wantErr: true,
		},
		{
			name: "nil Prometheus config",
			cfg: &config.OtelConfig{
				Prometheus: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := initMetricProvider(ctx, tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
