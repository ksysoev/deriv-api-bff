package repo

import (
	"fmt"
	"html/template"
	"slices"
	"sort"

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
	Method  string           `mapstructure:"method"`
	Params  validator.Config `mapstructure:"params"`
	Backend []BackendConfig  `mapstructure:"backend"`
}

type BackendConfig struct {
	FieldsMap       map[string]string `mapstructure:"fields_map"`
	DependsOn       []string          `mapstructure:"depends_on"`
	ResponseBody    string            `mapstructure:"response_body"`
	RequestTemplate string            `mapstructure:"request_template"`
	Allow           []string          `mapstructure:"allow"`
}

type CallsRepository struct {
	calls map[string]core.Handler
}

// NewCallsRepository creates a new instance of CallsRepository based on the provided configuration.
// It takes cfg of type *CallsConfig which contains the configuration for the calls.
// It returns a pointer to CallsRepository and an error if the repository creation fails.
// It returns an error if the validator creation fails or if there is an error parsing the request template.
func NewCallsRepository(cfg *CallsConfig) (*CallsRepository, error) {
	r := &CallsRepository{
		calls: make(map[string]core.Handler),
	}

	factory := func() handler.WaitComposer {
		return composer.New()
	}

	for _, call := range cfg.Calls {
		valid, err := validator.New(call.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to create validator: %w", err)
		}

		procs := make([]handler.RenderParser, 0, len(call.Backend))

		call.Backend, err = sortBackends(call.Backend)
		if err != nil {
			return nil, fmt.Errorf("failed to sort backends: %w", err)
		}

		for _, req := range call.Backend {
			tmplt, err := template.New("request").Parse(req.RequestTemplate)
			if err != nil {
				return nil, err
			}

			procs = append(procs, processor.New(&processor.Config{
				Tmplt:        tmplt,
				FieldMap:     req.FieldsMap,
				ResponseBody: req.ResponseBody,
				Allow:        req.Allow,
			}))
		}

		r.calls[call.Method] = handler.New(valid, procs, factory)
	}

	return r, nil
}

// GetCall retrieves a CallRunConfig based on the provided method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a pointer to a CallRunConfig if the method exists in the repository, otherwise it returns nil.
func (r *CallsRepository) GetCall(method string) core.Handler {
	return r.calls[method]
}

// sortBackends sorts a slice of BackendConfig based on their dependencies.
// It takes a slice of BackendConfig as input.
// It returns a sorted slice of BackendConfig and an error if a circular dependency is detected.
// It returns an error if there is a circular dependency among the backends.
func sortBackends(be []BackendConfig) ([]BackendConfig, error) {
	hasCircularDependency := false

	sort.Slice(be, func(i, j int) bool {
		dep1 := slices.Contains(be[i].DependsOn, be[j].ResponseBody)
		dep2 := slices.Contains(be[j].DependsOn, be[i].ResponseBody)

		if dep1 && dep2 {
			hasCircularDependency = true
			return false
		}

		if !dep1 && !dep2 {
			// Preserve order of backends that do not depend on each other
			return i < j
		}

		return !dep1
	})

	if hasCircularDependency {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return be, nil
}
