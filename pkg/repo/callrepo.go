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

	for _, call := range cfg.Calls {
		valid, err := validator.New(call.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to create validator: %w", err)
		}

		procs := make([]handler.RenderParser, 0, len(call.Backend))

		graph := createDepGraph(call.Backend)

		call.Backend, err = topSortDFS(call.Backend)
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

		factory := func(waiter core.Waiter) handler.WaitComposer {
			return composer.New(graph, waiter)
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

// topSortDFS performs a topological sort on a slice of BackendConfig using Depth-First Search (DFS).
// It takes a slice of BackendConfig as input and returns a sorted slice of BackendConfig and an error.
// It returns an error if a circular dependency is detected among the BackendConfig elements.
// Each BackendConfig element must have a unique ResponseBody and a list of dependencies specified in DependsOn.
func topSortDFS(be []BackendConfig) ([]BackendConfig, error) {
	graph := createDepGraph(be)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var sorted []BackendConfig

	var dfs func(string) error
	dfs = func(v string) error {
		if recStack[v] {
			return fmt.Errorf("circular dependency detected")
		}

		if visited[v] {
			return nil
		}

		visited[v] = true
		recStack[v] = true

		for _, u := range graph[v] {
			if err := dfs(u); err != nil {
				return err
			}
		}

		var indx int

		for i, b := range be {
			if b.ResponseBody == v {
				indx = i
				break
			}
		}

		recStack[v] = false

		sorted = append(sorted, be[indx])

		return nil
	}

	for k := range graph {
		if err := dfs(k); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

func createDepGraph(be []BackendConfig) map[string][]string {
	graph := make(map[string][]string)

	for _, b := range be {
		graph[b.ResponseBody] = b.DependsOn
	}

	return graph
}
