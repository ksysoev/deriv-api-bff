package processor

import (
	"html/template"
	"io"
)

type Processor interface {
	Name() string
	Render(w io.Writer, reqID int64, params, deps map[string]any) error
	Parse(data []byte) (resp, filetered map[string]any, err error)
}

type Config struct {
	Tmplt        *template.Template
	FieldMap     map[string]string
	ResponseBody string
	Allow        []string
}

func New(cfg *Config) Processor {
	return NewDeriv(cfg)
}
