package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServePrometheus(t *testing.T) {
	tests := []struct {
		cfg     *PrometheusConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &PrometheusConfig{
				Listen: ":0",
				Path:   "/metrics",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty listen",
			cfg: &PrometheusConfig{
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
		cfg     *PrometheusConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &PrometheusConfig{
				Listen: ":0",
				Path:   "/metrics",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty listen",
			cfg: &PrometheusConfig{
				Listen: "",
				Path:   "/metrics",
			},
			wantErr: true,
		},
		{
			name: "invalid config - empty path",
			cfg: &PrometheusConfig{
				Listen: ":0",
				Path:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid config - empty listen and path",
			cfg: &PrometheusConfig{
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
		cfg     *OtelConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config with Prometheus",
			cfg: &OtelConfig{
				Prometheus: &PrometheusConfig{
					Listen: ":0",
					Path:   "/metrics",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid Prometheus config - empty listen",
			cfg: &OtelConfig{
				Prometheus: &PrometheusConfig{
					Listen: "",
					Path:   "/metrics",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid Prometheus config - empty path",
			cfg: &OtelConfig{
				Prometheus: &PrometheusConfig{
					Listen: ":0",
					Path:   "",
				},
			},
			wantErr: true,
		},
		{
			name: "nil Prometheus config",
			cfg: &OtelConfig{
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
