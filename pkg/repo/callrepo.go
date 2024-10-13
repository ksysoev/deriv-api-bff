package repo

import (
	"fmt"
	"html/template"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/composer"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
)

type CallsConfig struct {
	Calls []CallConfig `mapstructure:"calls"`
}

type CallConfig struct {
	Method  string                    `mapstructure:"method"`
	Params  validator.ValidatorConfig `mapstructure:"params"`
	Backend []BackendConfig           `mapstructure:"backend"`
}

type BackendConfig struct {
	FieldsMap       map[string]string `mapstructure:"fields_map"`
	ResponseBody    string            `mapstructure:"response_body"`
	RequestTemplate string            `mapstructure:"request_template"`
	Allow           []string          `mapstructure:"allow"`
}

type CallsRepository struct {
	calls map[string]core.Handler
}

// NewCallsRepository initializes a new CallsRepository based on the provided CallsConfig.
// It takes cfg of type *CallsConfig which contains the configuration for the calls.
// It returns a pointer to CallsRepository and an error.
// It returns an error if there is an issue parsing the request templates.
// Edge cases include handling empty or nil CallsConfig, which will result in an empty CallsRepository.
func NewCallsRepository(cfg *CallsConfig) (*CallsRepository, error) {
	r := &CallsRepository{
		calls: make(map[string]core.Handler),
	}

	composerFactory := func() handler.Composer {
		return composer.New()
	}

	for _, call := range cfg.Calls {
		validator, err := validator.New(call.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to create validator: %w", err)
		}

		processors := make([]handler.Processor, 0, len(call.Backend))
		for _, req := range call.Backend {
			tmplt, err := template.New("request").Parse(req.RequestTemplate)
			if err != nil {
				return nil, err
			}

			p := processor.New(&processor.Config{
				Tmplt:        tmplt,
				FieldMap:     req.FieldsMap,
				ResponseBody: req.ResponseBody,
				Allow:        req.Allow,
			})

			processors = append(processors, p)
		}

		r.calls[call.Method] = handler.New(validator, processors, composerFactory)
	}

	return r, nil
}

// GetCall retrieves a CallRunConfig based on the provided method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a pointer to a CallRunConfig if the method exists in the repository, otherwise it returns nil.
func (r *CallsRepository) GetCall(method string) core.Handler {
	return r.calls[method]
}
