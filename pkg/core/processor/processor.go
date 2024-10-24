package processor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
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

type templateData struct {
	Params map[string]any
	Resp   map[string]any
	ReqID  int64
}

// New creates and returns a new Processor instance configured with the provided Config.
// It takes a single parameter cfg of type *Config which contains the necessary configuration.
// It returns a pointer to a Processor struct initialized with the values from the Config.
func New(cfg *Config) *Processor {
	return &Processor{
		tmpl:         cfg.Tmplt,
		fieldMap:     cfg.FieldMap,
		responseBody: cfg.ResponseBody,
		allow:        cfg.Allow,
	}
}

// Name returns the name of the Processor as a string.
// It does not take any parameters.
// It returns a string which is the response body of the Processor.
func (p *Processor) Name() string {
	return p.responseBody
}

// Render generates and writes the rendered template to the provided writer.
// It takes a writer w of type io.Writer, a request ID reqID of type int64,
// and two maps params and deps of type map[string]any.
// It returns an error if the template execution fails.
// If deps or params are nil, they are initialized as empty maps before template execution.
func (p *Processor) Render(w io.Writer, reqID int64, params, deps map[string]any) error {
	if deps == nil {
		deps = make(map[string]any)
	}

	if params == nil {
		params = make(map[string]any)
	}

	data := templateData{
		Params: params,
		ReqID:  reqID,
		Resp:   deps,
	}

	return p.tmpl.Execute(w, data)
}

// Parse processes the input data and filters the response based on allowed keys.
// It takes data of type []byte.
// It returns three values: resp which is a map[string]any containing the parsed response,
// filetered which is a map[string]any containing the filtered response based on allowed keys,
// and an error if the parsing fails.
// It returns an error if the input data cannot be parsed.
// Edge case: If the response does not contain an expected key, a warning is logged and the key is skipped.
func (p *Processor) Parse(data []byte) (resp, filetered map[string]any, err error) {
	resp, err = p.parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to parse response %s: %w", p.responseBody, err)
	}

	filetered = make(map[string]any, len(p.allow))

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

		filetered[destKey] = resp[key]
	}

	return resp, filetered, nil
}

// parse unmarshals JSON data into a map and processes the response body.
// It takes data of type []byte and returns a map[string]any and an error.
// It returns an error if the JSON unmarshalling fails, if the response contains an error,
// or if the response body is not found or is in an unexpected format.
// If the response body is a map, it returns it directly. If it is a list, it wraps it in a map with the key "list".
// If it is any other type, it wraps it in a map with the key "value".
func (p *Processor) parse(data []byte) (map[string]any, error) {
	var rdata map[string]any

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, err
	}

	if errData, ok := rdata["error"]; ok {
		return nil, NewAPIError(errData)
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
