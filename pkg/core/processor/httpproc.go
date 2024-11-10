package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/core/tmpl"
)

type HTTPProc struct {
	urlTemplate *tmpl.URLTmpl
	tmpl        *tmpl.Tmpl
	fieldMap    map[string]string
	headers     map[string]string
	name        string
	method      string
	allow       []string
}

// NewHTTP creates a new instance of HTTPProc based on the provided configuration.
// It takes a single parameter cfg of type *Config which contains the necessary configuration details.
// It returns a pointer to an HTTPProc initialized with the values from the configuration.
func NewHTTP(cfg *Config) (*HTTPProc, error) {
	rawTmpl, err := json.Marshal(cfg.Tmplt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request template: %w", err)
	}

	reqTmpl, err := tmpl.New(string(rawTmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	urlTmpl, err := tmpl.NewURLTmpl(cfg.URLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL template: %w", err)
	}

	return &HTTPProc{
		name:        cfg.Name,
		method:      cfg.Method,
		urlTemplate: urlTmpl,
		tmpl:        reqTmpl,
		fieldMap:    cfg.FieldMap,
		allow:       cfg.Allow,
		headers:     cfg.Headers,
	}, nil
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
func (p *HTTPProc) Render(ctx context.Context, reqID string, param, deps map[string]any) (core.Request, error) {
	data := templateData{
		Params: param,
		Resp:   deps,
		ReqID:  reqID,
	}

	url, err := p.urlTemplate.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("fail to execute URL template %s: %w", p.name, err)
	}

	var body []byte

	if p.tmpl != nil {
		body, err = p.tmpl.Execute(data)
		if err != nil {
			return nil, fmt.Errorf("fail to execute request template %s: %w", p.name, err)
		}
	}

	req := request.NewHTTPReq(ctx, p.method, url, body, reqID)

	if p.headers == nil {
		for key, value := range p.headers {
			req.AddHeader(key, value)
		}
	}

	return req, nil
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
		if errData, ok := respBody["error"]; ok {
			return nil, NewAPIError(errData)
		}

		return respBody, nil
	case []any:
		return map[string]any{"list": respBody}, nil
	case any:
		return map[string]any{"value": respBody}, nil
	default:
		return nil, fmt.Errorf("response body is in unexpected format")
	}
}
