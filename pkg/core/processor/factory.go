package processor

import (
	"fmt"
	"html/template"
	"io"
)

type Processor interface {
	Name() string
	Render(w io.Writer, reqID int64, params, deps map[string]any) error
	Parse(data []byte) (resp, filetered map[string]any, err error)
}

type Config struct {
	Name         string
	Method       string
	URLTemplate  string
	Tmplt        *template.Template
	FieldMap     map[string]string
	ResponseBody string
	Allow        []string
}

func New(cfg *Config) (Processor, error) {
	switch {
	case isDerivConfig(cfg) && isHTTPConfig(cfg):
		return nil, fmt.Errorf("ambiguous processor configuration")
	case isDerivConfig(cfg):
		return NewDeriv(cfg), nil
	case isHTTPConfig(cfg):
		return NewHTTP(cfg), nil
	default:
		return nil, fmt.Errorf("invalid processor configuration")
	}
}

func isDerivConfig(cfg *Config) bool {
	return cfg.ResponseBody != ""
}

func isHTTPConfig(cfg *Config) bool {
	return cfg.Method != "" && cfg.URLTemplate != ""
}
