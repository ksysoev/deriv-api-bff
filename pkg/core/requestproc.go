package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

type RequestProcessor struct {
	tempate      *template.Template
	fieldsMap    map[string]string
	responseBody string
	allow        []string
	mu           sync.Mutex
}

// Render generates a byte slice from the provided TemplateData using the RequestProcessor's template.
// It takes data of type TemplateData.
// It returns a byte slice containing the rendered template and an error if the template execution fails.
// It returns an error if the template execution encounters an issue.
func (r *RequestProcessor) Render(data handler.TemplateData) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var buf bytes.Buffer
	if err := r.tempate.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ParseResp parses the given JSON byte data into a map[string]any.
// It takes data of type []byte and returns a map[string]any and an error.
// It returns an error if the JSON unmarshalling fails, if the response contains an "error" key,
// or if the expected response body is not found or is in an unexpected format.
// If the response body is a map, it returns the map directly.
// If the response body is a list, it returns a map with the key "list" containing the list.
// If the response body is a single value, it returns a map with the key "value" containing the value.
func (r *RequestProcessor) ParseResp(data []byte) (map[string]any, error) {
	var rdata map[string]any

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, err
	}

	if _, ok := rdata["error"]; ok {
		return nil, NewAPIError(r.responseBody, rdata)
	}

	rb, ok := rdata[r.responseBody]
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
