package processor

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/response"
)

type Processor interface {
	Name() string
	Render(ctx context.Context, reqID string, params []byte, deps map[string]any) (core.Request, error)
	Parse(data []byte) (*response.Response, error)
}

type Config struct {
	Request   map[string]any    `json:"request,omitempty" yaml:"request,omitempty"`
	FieldMap  map[string]string `json:"fields_map,omitempty" yaml:"fields_map,omitempty"`
	Headers   map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	Method    string            `json:"method,omitempty" yaml:"method,omitempty"`
	URL       string            `json:"url,omitempty" yaml:"url,omitempty"`
	DependsOn []string          `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Allow     []string          `json:"allow,omitempty" yaml:"allow,omitempty"`
}

// New creates a new Processor based on the provided configuration.
// It takes cfg of type *Config.
// It returns a Processor and an error.
// It returns an error if the configuration is ambiguous or invalid.
func New(cfg *Config) (Processor, error) {
	switch {
	case isHTTPConfig(cfg):
		return NewHTTP(cfg)
	case isDerivConfig(cfg):
		return NewDeriv(cfg)
	default:
		return nil, fmt.Errorf("invalid processor configuration")
	}
}

// isDerivConfig checks if the given configuration is a Deriv configuration.
// It takes a single parameter cfg of type *Config.
// It returns a boolean value indicating whether the ResponseBody field of the configuration is not empty.
func isDerivConfig(cfg *Config) bool {
	return len(cfg.Request) > 0 && cfg.Method == "" && cfg.URL == ""
}

// isHTTPConfig checks if the given configuration is for an HTTP request.
// It takes a single parameter cfg of type *Config.
// It returns a boolean value: true if both Method and URLTemplate fields of cfg are non-empty, otherwise false.
func isHTTPConfig(cfg *Config) bool {
	return cfg.Method != "" && cfg.URL != ""
}
