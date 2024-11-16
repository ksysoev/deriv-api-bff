package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/core/response"
	"github.com/ksysoev/deriv-api-bff/pkg/core/tmpl"
)

type DerivProc struct {
	name     string
	tmpl     *tmpl.Tmpl
	fieldMap map[string]string
	allow    []string
}

type templateData struct {
	Resp   map[string]any  `json:"resp"`
	ReqID  string          `json:"req_id"`
	Params json.RawMessage `json:"params"`
}

type passthrough struct {
	ReqID string `json:"req_id"`
}

// NewDeriv creates and returns a new Processor instance configured with the provided Config.
// It takes a single parameter cfg of type *Config which contains the necessary configuration.
// It returns a pointer to a Processor struct initialized with the values from the Config.
func NewDeriv(cfg *Config) (*DerivProc, error) {
	t := cfg.Request
	t["passthrough"] = passthrough{ReqID: "${req_id}"}

	rawTmpl, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request template: %w", err)
	}

	reqTmpl, err := tmpl.New(string(rawTmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &DerivProc{
		name:     cfg.Name,
		tmpl:     reqTmpl,
		fieldMap: cfg.FieldMap,
		allow:    cfg.Allow,
	}, nil
}

// Name returns the name of the Processor as a string.
// It does not take any parameters.
// It returns a string which is the response body of the Processor.
func (p *DerivProc) Name() string {
	return p.name
}

// Render generates and writes the rendered template to the provided writer.
// It takes a writer w of type io.Writer, a request ID reqID of type int64,
// and two maps params and deps of type map[string]any.
// It returns an error if the template execution fails.
// If deps or params are nil, they are initialized as empty maps before template execution.
func (p *DerivProc) Render(ctx context.Context, reqID string, params []byte, deps map[string]any) (core.Request, error) {
	if deps == nil {
		deps = make(map[string]any)
	}

	if params == nil {
		params = []byte("{}")
	}

	data := templateData{
		Params: params,
		ReqID:  reqID,
		Resp:   deps,
	}

	req, err := p.tmpl.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return request.NewRequest(ctx, request.TextMessage, req), nil
}

// Parse processes the input data and filters the response based on allowed keys.
// It takes data of type []byte.
// It returns three values: resp which is a map[string]any containing the parsed response,
// filetered which is a map[string]any containing the filtered response based on allowed keys,
// and an error if the parsing fails.
// It returns an error if the input data cannot be parsed.
// Edge case: If the response does not contain an expected key, a warning is logged and the key is skipped.
func (p *DerivProc) Parse(data []byte) (*response.Response, error) {
	resp, err := p.parse(data)
	if err != nil {
		return nil, fmt.Errorf("fail to parse response %s: %w", p.name, err)
	}

	filetered := make(map[string]json.RawMessage, len(p.allow))

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

	return response.New(resp, filetered), nil
}

// parse unmarshals JSON data into a map and processes the response body.
// It takes data of type []byte and returns a map[string]any and an error.
// It returns an error if the JSON unmarshalling fails, if the response contains an error,
// or if the response body is not found or is in an unexpected format.
// If the response body is a map, it returns it directly. If it is a list, it wraps it in a map with the key "list".
// If it is any other type, it wraps it in a map with the key "value".
func (p *DerivProc) parse(data []byte) (map[string]json.RawMessage, error) {
	var rdata map[string]json.RawMessage

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if errData, ok := rdata["error"]; ok {
		return nil, NewAPIError(errData)
	}

	msgTypeRaw, ok := rdata["msg_type"]
	if !ok {
		return nil, fmt.Errorf("msg_type not found or not a string")
	}

	if len(msgTypeRaw) < 3 || msgTypeRaw[0] != '"' || msgTypeRaw[len(msgTypeRaw)-1] != '"' {
		return nil, fmt.Errorf("msg_type not a valid string")
	}

	msgType := msgTypeRaw[1 : len(msgTypeRaw)-1]

	rb, ok := rdata[string(msgType)]
	if !ok {
		return nil, fmt.Errorf("response body not found")
	}

	switch rb[0] {
	case '{':
		var respBody map[string]json.RawMessage

		if err := json.Unmarshal(rb, &respBody); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}

		return respBody, nil
	case '[':
		return map[string]json.RawMessage{"list": rb}, nil
	default:
		return map[string]json.RawMessage{"value": rb}, nil
	}
}
