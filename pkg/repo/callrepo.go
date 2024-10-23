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

type EtcdConfig struct {
	Servers            []string `mapstructure:"servers"`
	DialTimeoutSeconds int      `mapstructure:"dialTimeoutSeconds"`
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
		err := createHandler(call, r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func createHandler(call CallConfig, r *CallsRepository) error {
	valid, err := validator.New(call.Params)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	procs := make([]handler.RenderParser, 0, len(call.Backend))
	graph := createDepGraph(call.Backend)

	call.Backend, err = topSortDFS(call.Backend)
	if err != nil {
		return fmt.Errorf("failed to sort backends: %w", err)
	}

	for _, req := range call.Backend {
		tmplt, err := template.New("request").Parse(req.RequestTemplate)
		if err != nil {
			return err
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

	return nil
}

// GetCall retrieves a CallRunConfig based on the provided method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a pointer to a CallRunConfig if the method exists in the repository, otherwise it returns nil.
func (r *CallsRepository) GetCall(method string) core.Handler {
	return r.calls[method]
}

// UpdateCalls refreshes the CallsConfig and rebuilds the handlers for calls accordingly
// It takes a single parameter that is pointer to the new calls config.
// The current implementation will completely overwrite the old config with new config.
// It returns an error to if any while building the handlers, otherwise it returns nil.
func (r *CallsRepository) UpdateCalls(calls *CallsConfig) error {
	r.calls = make(map[string]core.Handler)

	for _, call := range calls.Calls {
		err := createHandler(call, r)
		if err != nil {
			return err
		}
	}

	return nil
}

// topSortDFS performs a topological sort on a slice of BackendConfig using Depth-First Search (DFS).
// It takes a slice of BackendConfig as input and returns a sorted slice of BackendConfig and an error.
// It returns an error if a circular dependency is detected among the BackendConfig elements.
// Each BackendConfig element must have a unique ResponseBody and a list of dependencies specified in DependsOn.
func topSortDFS(be []BackendConfig) ([]BackendConfig, error) {
	graph := createDepGraph(be)
	visited := make(map[string]bool, len(be))
	recStack := make(map[string]bool, len(be))
	sorted := make([]BackendConfig, 0, len(be))

	indexMap := make(map[string]int, len(be))
	for i, b := range be {
		indexMap[b.ResponseBody] = i
	}

	var dfs func(string) error

	dfs = func(v string) error {
		if recStack[v] {
			return fmt.Errorf("circular dependency detected at %s", v)
		}

		if visited[v] {
			return nil
		}

		visited[v], recStack[v] = true, true

		defer func() { recStack[v] = false }() // Ensure recStack is reset

		for _, u := range graph[v] {
			if err := dfs(u); err != nil {
				return err
			}
		}

		sorted = append(sorted, be[indexMap[v]])

		return nil
	}

	for _, b := range be {
		if err := dfs(b.ResponseBody); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

// createDepGraph constructs a dependency graph from a slice of BackendConfig.
// It takes a single parameter be which is a slice of BackendConfig.
// It returns a map where the keys are response bodies and the values are slices of dependencies.
func createDepGraph(be []BackendConfig) map[string][]string {
	graph := make(map[string][]string)

	for _, b := range be {
		graph[b.ResponseBody] = b.DependsOn
	}

	return graph
}
