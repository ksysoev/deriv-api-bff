package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sync"
)

type RequestProcessor struct {
	tempate      *template.Template
	allow        []string
	responseBody string
	mu           sync.Mutex
}

func (r *RequestProcessor) Render(data TemplateData) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var buf bytes.Buffer
	err := r.tempate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

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
