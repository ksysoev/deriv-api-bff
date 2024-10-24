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
				ResponseBody: "some response",
			},
			wantErr: false,
		},
		{
			name: "Valid HTTP Config",
			cfg: &Config{
				Method:      "GET",
				URLTemplate: "http://example.com",
			},
			wantErr: false,
		},
		{
			name: "Ambiguous Config",
			cfg: &Config{
				Method:       "GET",
				URLTemplate:  "http://example.com",
				ResponseBody: "some response",
			},
			wantErr: true,
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
				ResponseBody: "some response",
			},
			want: true,
		},
		{
			name: "Non-Deriv Config",
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
				Method:      "GET",
				URLTemplate: "http://example.com",
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
