package processor

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type Processor interface {
	Name() string
	Render(ctx context.Context, reqID string, params map[string]any, deps map[string]any) (core.Request, error)
	Parse(data []byte) (resp, filetered map[string]any, err error)
}

type Config struct {
	Name         string            `json:"name" yaml:"name"`
	Method       string            `json:"method" yaml:"method"`
	URLTemplate  string            `json:"url_template" yaml:"url_template"`
	Tmplt        map[string]any    `json:"request_template" yaml:"request_template"`
	FieldMap     map[string]string `json:"fields_map" yaml:"fields_map"`
	DependsOn    []string          `json:"depends_on" yaml:"depends_on"`
	ResponseBody string            `json:"response_body" yaml:"response_body"`
	Allow        []string          `json:"allow" yaml:"allow"`
}

// New creates a new Processor based on the provided configuration.
// It takes cfg of type *Config.
// It returns a Processor and an error.
// It returns an error if the configuration is ambiguous or invalid.
func New(cfg *Config) (Processor, error) {
	switch {
	case isDerivConfig(cfg) && isHTTPConfig(cfg):
		return nil, fmt.Errorf("ambiguous processor configuration")
	case isDerivConfig(cfg):
		return NewDeriv(cfg)
	case isHTTPConfig(cfg):
		return NewHTTP(cfg)
	default:
		return nil, fmt.Errorf("invalid processor configuration")
	}
}

// isDerivConfig checks if the given configuration is a Deriv configuration.
// It takes a single parameter cfg of type *Config.
// It returns a boolean value indicating whether the ResponseBody field of the configuration is not empty.
func isDerivConfig(cfg *Config) bool {
	return cfg.ResponseBody != "" && len(cfg.Tmplt) > 0
}

// isHTTPConfig checks if the given configuration is for an HTTP request.
// It takes a single parameter cfg of type *Config.
// It returns a boolean value: true if both Method and URLTemplate fields of cfg are non-empty, otherwise false.
func isHTTPConfig(cfg *Config) bool {
	return cfg.Method != "" && cfg.URLTemplate != ""
}
