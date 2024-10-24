package processor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
)

type HTTPProc struct {
	name        string
	method      string
	urlTemplate string
	tmpl        *template.Template
	fieldMap    map[string]string
	allow       []string
}

func NewHTTP(cfg *Config) *HTTPProc {
	return &HTTPProc{
		name:        cfg.Name,
		method:      cfg.Method,
		urlTemplate: cfg.URLTemplate,
		tmpl:        cfg.Tmplt,
		fieldMap:    cfg.FieldMap,
		allow:       cfg.Allow,
	}
}

func (p *HTTPProc) Name() string {
	return p.name
}

func (p *HTTPProc) Render(w io.Writer, reqID int64, params, deps map[string]any) error {
	return fmt.Errorf("http processor not implemented")
}

func (p *HTTPProc) Parse(data []byte) (resp, filetered map[string]any, err error) {
	resp, err = p.parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to parse response %s: %w", p.name, err)
	}

	filetered = make(map[string]any, len(p.allow))

	for _, key := range p.allow {
		if _, ok := resp[key]; !ok {
			slog.Warn("Response body does not contain expeted key", slog.String("key", key), slog.String("name", p.name))
			continue
		}

		destKey := key

		if p.fieldMap != nil {
			if mappedKey, ok := p.fieldMap[key]; ok {
				destKey = mappedKey
			}
		}

		filetered[destKey] = resp[key]
	}

	return resp, filetered, nil
}

func (p *HTTPProc) parse(data []byte) (map[string]any, error) {
	var rdata any

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, err
	}

	switch respBody := rdata.(type) {
	case map[string]any:
		return respBody, nil
	case []any:
		return map[string]any{"list": respBody}, nil
	case any:
		return map[string]any{"value": respBody}, nil
	default:
		return nil, fmt.Errorf("response body is in unexpected format")
	}
}
