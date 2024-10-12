package processor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

type Config struct {
	Tmplt        *template.Template
	FieldMap     map[string]string
	ResponseBody string
	Allow        []string
}

type Processor struct {
	tmpl         *template.Template
	fieldMap     map[string]string
	responseBody string
	allow        []string
}

func New(cfg *Config) *Processor {
	return &Processor{
		tmpl:         cfg.Tmplt,
		fieldMap:     cfg.FieldMap,
		responseBody: cfg.ResponseBody,
		allow:        cfg.Allow,
	}
}

// Render generates a byte slice from the provided TemplateData using the RequestProcessor's template.
// It takes data of type TemplateData.
// It returns a byte slice containing the rendered template and an error if the template execution fails.
// It returns an error if the template execution encounters an issue.
func (p *Processor) Render(w io.Writer, data handler.TemplateData) error {
	return p.tmpl.Execute(w, data)
}

// ParseResp parses the given JSON byte data into a map[string]any.
// It takes data of type []byte and returns a map[string]any and an error.
// It returns an error if the JSON unmarshalling fails, if the response contains an "error" key,
// or if the expected response body is not found or is in an unexpected format.
// If the response body is a map, it returns the map directly.
// If the response body is a list, it returns a map with the key "list" containing the list.
// If the response body is a single value, it returns a map with the key "value" containing the value.
func (p *Processor) Parse(data []byte) (map[string]any, error) {
	resp, err := p.parse(data)
	if err != nil {
		return nil, fmt.Errorf("fail to parse response %s: %w", p.responseBody, err)
	}

	result := make(map[string]any, len(p.allow))

	for _, key := range p.allow {
		if _, ok := resp[key]; !ok {
			slog.Warn("Response body does not contain expeted key", slog.String("key", key), slog.String("response_body", p.responseBody))
			continue
		}

		destKey := key

		if p.fieldMap != nil {
			if mappedKey, ok := p.fieldMap[key]; ok {
				destKey = mappedKey
			}
		}

		result[destKey] = resp[key]
	}

	return result, nil
}

// ParseResp parses the given JSON byte data into a map[string]any.
// It takes data of type []byte and returns a map[string]any and an error.
// It returns an error if the JSON unmarshalling fails, if the response contains an "error" key,
// or if the expected response body is not found or is in an unexpected format.
// If the response body is a map, it returns the map directly.
// If the response body is a list, it returns a map with the key "list" containing the list.
// If the response body is a single value, it returns a map with the key "value" containing the value.
func (p *Processor) parse(data []byte) (map[string]any, error) {
	var rdata map[string]any

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, err
	}

	if _, ok := rdata["error"]; ok {
		return nil, NewAPIError(p.responseBody, rdata)
	}

	rb, ok := rdata[p.responseBody]
	if !ok {
		return nil, fmt.Errorf("response body not found")
	}

	switch respBody := rb.(type) {
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
