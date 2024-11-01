package repo

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/composer"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/mitchellh/mapstructure"
)

type CallsRepository struct {
	mu            *sync.Mutex
	calls         map[string]core.Handler
	onUpdateEvent *config.Event[any]
}

// NewCallsRepository creates a new instance of CallsRepository based on the provided configuration.
// It takes cfg of type *CallsConfig which contains the configuration for the calls.
// It returns a pointer to CallsRepository and an error if the repository creation fails.
// It returns an error if the validator creation fails or if there is an error parsing the request template.
func NewCallsRepository(cfg *config.CallsConfig, event *config.Event[any]) (*CallsRepository, error) {
	handlerMap := make(map[string]core.Handler)

	for _, call := range cfg.Calls {
		err := createHandler(call, handlerMap)
		if err != nil {
			return nil, err
		}
	}

	r := &CallsRepository{
		calls:         handlerMap,
		mu:            &sync.Mutex{},
		onUpdateEvent: event,
	}

	r.onUpdateEvent.RegisterHandler(func(_ context.Context, cc any) {
		ccMap, ok := cc.(map[string]any)

		if !ok {
			slog.Error("Error while trying to update calls config: incoming config is not of type `map[any]`")
			return
		}

		r.UpdateCalls(ccMap)
	})

	return r, nil
}

func createHandler(call config.CallConfig, handlerMap map[string]core.Handler) error {
	valid, err := validator.New(call.Params)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	for _, req := range call.Backend {
		if req.Name == "" {
			req.Name = req.ResponseBody
		}

		if req.Name == "" {
			return fmt.Errorf("name or response_body must be provided")
		}
	}

	procs := make([]handler.RenderParser, 0, len(call.Backend))
	graph := createDepGraph(call.Backend)

	call.Backend, err = topSortDFS(call.Backend)
	if err != nil {
		return fmt.Errorf("failed to sort backends: %w", err)
	}

	for _, req := range call.Backend {
		p, err := processor.New(&processor.Config{
			Name:         req.Name,
			Tmplt:        req.RequestTemplate,
			FieldMap:     req.FieldsMap,
			ResponseBody: req.ResponseBody,
			Allow:        req.Allow,
			Method:       req.Method,
			URLTemplate:  req.URLTemplate,
		})

		if err != nil {
			return fmt.Errorf("failed to create processor: %w", err)
		}

		procs = append(procs, p)
	}

	factory := func(waiter core.Waiter) handler.WaitComposer {
		return composer.New(graph, waiter)
	}

	handlerMap[call.Method] = handler.New(valid, procs, factory)

	return nil
}

// GetCall retrieves a CallRunConfig based on the provided method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a pointer to a CallRunConfig if the method exists in the repository, otherwise it returns nil.
func (r *CallsRepository) GetCall(method string) core.Handler {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls[method]
}

// UpdateCalls refreshes the CallsConfig and rebuilds the handlers for calls accordingly
// It takes a single parameter that is pointer to the new calls config.
// The current implementation will completely overwrite the old config with new config.
// It returns an error to if any while building the handlers, otherwise it returns nil.
func (r *CallsRepository) UpdateCalls(callsMap map[string]any) {
	newHandlerMap := make(map[string]core.Handler)
	calls := &config.CallsConfig{}

	err := mapstructure.Decode(&callsMap, calls)

	if err != nil {
		slog.Warn(fmt.Sprintf("Error while decoding calls config map: %v", err))
		return
	}

	for _, call := range calls.Calls {
		err := createHandler(call, newHandlerMap)
		if err != nil {
			slog.Error(fmt.Sprintf("Error while updating calls config: %v", err))
			return
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = newHandlerMap
}

// topSortDFS performs a topological sort on a slice of BackendConfig using Depth-First Search (DFS).
// It takes a slice of BackendConfig as input and returns a sorted slice of BackendConfig and an error.
// It returns an error if a circular dependency is detected among the BackendConfig elements.
// Each BackendConfig element must have a unique ResponseBody and a list of dependencies specified in DependsOn.
func topSortDFS(be []*config.BackendConfig) ([]*config.BackendConfig, error) {
	graph := createDepGraph(be)
	visited := make(map[string]bool, len(be))
	recStack := make(map[string]bool, len(be))
	sorted := make([]*config.BackendConfig, 0, len(be))

	indexMap := make(map[string]int, len(be))
	for i, b := range be {
		if _, ok := indexMap[b.Name]; ok {
			return nil, fmt.Errorf("duplicate backend name: %s", b.Name)
		}

		indexMap[b.Name] = i
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
		if err := dfs(b.Name); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

// createDepGraph constructs a dependency graph from a slice of BackendConfig.
// It takes a single parameter be which is a slice of BackendConfig.
// It returns a map where the keys are response bodies and the values are slices of dependencies.
func createDepGraph(be []*config.BackendConfig) map[string][]string {
	graph := make(map[string][]string)

	for _, b := range be {
		graph[b.Name] = b.DependsOn
	}

	return graph
}
