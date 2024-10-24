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

// NewHTTP creates a new instance of HTTPProc based on the provided configuration.
// It takes a single parameter cfg of type *Config which contains the necessary configuration details.
// It returns a pointer to an HTTPProc initialized with the values from the configuration.
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

// Name returns the name of the HTTPProc instance.
// It takes no parameters.
// It returns a string which is the name of the HTTPProc instance.
func (p *HTTPProc) Name() string {
	return p.name
}

// Render processes the HTTP request and writes the response.
// It takes an io.Writer, an int64, and two maps of string to any type as parameters.
// It returns an error indicating that the HTTP processor is not implemented.
func (p *HTTPProc) Render(_ io.Writer, _ int64, _, _ map[string]any) error {
	return fmt.Errorf("http processor not implemented")
}

// Parse processes the input data and filters the response based on allowed keys.
// It takes data of type []byte and returns two maps: resp and filetered, both of type map[string]any, and an error if parsing fails.
// It returns an error if the input data cannot be parsed.
// Edge cases include missing expected keys in the response, which are logged as warnings.
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

// parse parses the given JSON data and returns it as a map[string]any.
// It takes a single parameter data of type []byte which is the JSON data to be parsed.
// It returns a map[string]any representing the parsed JSON data and an error if the parsing fails.
// It returns an error if the JSON data is malformed or if the response body is in an unexpected format.
// If the JSON data is an array, it wraps it in a map with the key "list".
// If the JSON data is a single value, it wraps it in a map with the key "value".
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
