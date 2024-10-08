package repo

import (
	"html/template"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type CallsConfig struct {
	Calls []CallConfig `mapstructure:"calls"`
}

type CallConfig struct {
	Method  string            `mapstructure:"method"`
	Params  map[string]string `mapstructure:"params"`
	Backend []BackendConfig   `mapstructure:"backend"`
}

type BackendConfig struct {
	FieldsMap       map[string]string `mapstructure:"fields_map"`
	ResponseBody    string            `mapstructure:"response_body"`
	RequestTemplate string            `mapstructure:"request_template"`
	Allow           []string          `mapstructure:"allow"`
}

type CallsRepository struct {
	calls map[string]*core.CallRunConfig
}

// NewCallsRepository initializes a new CallsRepository based on the provided CallsConfig.
// It takes cfg of type *CallsConfig which contains the configuration for the calls.
// It returns a pointer to CallsRepository and an error.
// It returns an error if there is an issue parsing the request templates.
// Edge cases include handling empty or nil CallsConfig, which will result in an empty CallsRepository.
func NewCallsRepository(cfg *CallsConfig) (*CallsRepository, error) {
	r := &CallsRepository{
		calls: make(map[string]*core.CallRunConfig),
	}

	for _, call := range cfg.Calls {
		c := &core.CallRunConfig{
			Requests: make(map[string]*core.RequestRunConfig),
		}

		for _, req := range call.Backend {
			tmplt, err := template.New("request").Parse(req.RequestTemplate)
			if err != nil {
				return nil, err
			}

			c.Requests[req.ResponseBody] = &core.RequestRunConfig{
				Tmplt:        tmplt,
				Allow:        req.Allow,
				FieldMap:     req.FieldsMap,
				ResponseBody: req.ResponseBody,
			}
		}

		r.calls[call.Method] = c
	}

	return r, nil
}

func (r *CallsRepository) GetCall(method string) *core.CallRunConfig {
	return r.calls[method]
}
