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

// Parse processes the input data and returns a parsed and filtered response.
// It takes data of type []byte.
// It returns a pointer to response.Response and an error.
// It returns an error if parsing or preparing the response fails.
func (p *DerivProc) Parse(data []byte) (*response.Response, error) {
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

// parse parses the given JSON data and extracts the relevant message body.
// It takes data of type []byte which is the JSON data to be parsed.
// It returns a json.RawMessage containing the extracted message body and an error if any occurs.
// It returns an error if the JSON data cannot be unmarshaled, if the "error" field is present in the data,
// if the "msg_type" field is missing or not a valid string, or if the response body corresponding to the msg_type is not found.
func (p *DerivProc) parse(data []byte) (json.RawMessage, error) {
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

	return rb, nil
}
