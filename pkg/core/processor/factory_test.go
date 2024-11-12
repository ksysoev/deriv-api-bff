package processor

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		cfg     *Config
		name    string
		wantErr bool
	}{
		{
			name: "Valid Deriv Config",
			cfg: &Config{
				Request: map[string]any{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Valid HTTP Config",
			cfg: &Config{
				Method: "GET",
				URL:    "/test/url",
			},
			wantErr: false,
		},
		{
			name:    "Invalid Config",
			cfg:     &Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsDerivConfig(t *testing.T) {
	tests := []struct {
		cfg  *Config
		name string
		want bool
	}{
		{
			name: "Deriv Config",
			cfg: &Config{
				Request: map[string]any{"key": "value"},
			},
			want: true,
		},
		{
			name: "Non-Deriv Config1",
			cfg:  &Config{},
			want: false,
		},
		{
			name: "Non-Deriv Config2",
			cfg:  &Config{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDerivConfig(tt.cfg); got != tt.want {
				t.Errorf("isDerivConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHTTPConfig(t *testing.T) {
	tests := []struct {
		cfg  *Config
		name string
		want bool
	}{
		{
			name: "HTTP Config",
			cfg: &Config{
				Method: "GET",
				URL:    "/test/url",
			},
			want: true,
		},
		{
			name: "Non-HTTP Config",
			cfg:  &Config{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHTTPConfig(tt.cfg); got != tt.want {
				t.Errorf("isHTTPConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
