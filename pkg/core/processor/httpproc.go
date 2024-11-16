package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/core/response"
	"github.com/ksysoev/deriv-api-bff/pkg/core/tmpl"
)

type errData struct {
	Err json.RawMessage `json:"error,omitempty"`
}

type HTTPProc struct {
	urlTemplate *tmpl.URLTmpl
	tmpl        *tmpl.Tmpl
	fieldMap    map[string]string
	headers     map[string]*tmpl.StrTmpl
	name        string
	method      string
	allow       []string
}

// NewHTTP creates a new instance of HTTPProc based on the provided configuration.
// It takes a single parameter cfg of type *Config which contains the necessary configuration details.
// It returns a pointer to an HTTPProc initialized with the values from the configuration.
func NewHTTP(cfg *Config) (*HTTPProc, error) {
	rawTmpl, err := json.Marshal(cfg.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request template: %w", err)
	}

	reqTmpl, err := tmpl.New(string(rawTmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	urlTmpl, err := tmpl.NewURLTmpl(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL template: %w", err)
	}

	headers := make(map[string]*tmpl.StrTmpl, len(cfg.Headers))

	for key, value := range cfg.Headers {
		t, err := tmpl.NewStrTmpl(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse header template %s: %w", key, err)
		}

		headers[key] = t
	}

	return &HTTPProc{
		name:        cfg.Name,
		method:      cfg.Method,
		urlTemplate: urlTmpl,
		tmpl:        reqTmpl,
		fieldMap:    cfg.FieldMap,
		allow:       cfg.Allow,
		headers:     headers,
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
func (p *HTTPProc) Render(ctx context.Context, reqID string, param []byte, deps map[string]any) (core.Request, error) {
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

	if p.headers != nil {
		for key, value := range p.headers {
			headerValue, err := value.Execute(data)
			if err != nil {
				return nil, fmt.Errorf("fail to execute header template %s: %w", key, err)
			}

			req.AddHeader(key, headerValue)
		}
	}

	return req, nil
}

// Parse processes the input data and filters the response based on allowed keys.
// It takes data of type []byte and returns two maps: resp and filetered, both of type map[string]any, and an error if parsing fails.
// It returns an error if the input data cannot be parsed.
// Edge cases include missing expected keys in the response, which are logged as warnings.
func (p *HTTPProc) Parse(data []byte) (*response.Response, error) {
	resp, err := p.parse(data)
	if err != nil {
		return nil, fmt.Errorf("fail to parse response %s: %w", p.name, err)
	}

	prepared, err := prepareResp(resp)
	if err != nil {
		return nil, fmt.Errorf("fail to prepare response %s: %w", p.name, err)
	}

	filetered := filterResp(prepared, p.allow, p.fieldMap)

	return response.New(resp, filetered), nil
}

// parse parses the given JSON data and returns it as a map[string]any.
// It takes a single parameter data of type []byte which is the JSON data to be parsed.
// It returns a map[string]any representing the parsed JSON data and an error if the parsing fails.
// It returns an error if the JSON data is malformed or if the response body is in an unexpected format.
// If the JSON data is an array, it wraps it in a map with the key "list".
// If the JSON data is a single value, it wraps it in a map with the key "value".
func (p *HTTPProc) parse(data []byte) (json.RawMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("response body not found")
	}

	if data[0] == '{' {
		var errRaw errData

		if err := json.Unmarshal(data, &errRaw); err != nil {
			return nil, fmt.Errorf("Fail to parse response: %s", err)
		}

		if errRaw.Err != nil {
			return nil, NewAPIError(errRaw.Err)
		}
	}

	return data, nil
}
