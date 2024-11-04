package handlerfactory

import (
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/composer"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
)

type Config struct {
	Method  string              `mapstructure:"method" yaml:"method"`
	Params  validator.Config    `mapstructure:"params" yaml:"params"`
	Backend []*processor.Config `mapstructure:"backend" yaml:"backend"`
}

func New(cfg Config) (string, core.Handler, error) {
	valid, err := validator.New(cfg.Params)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create validator: %w", err)
	}

	if cfg.Method == "" {
		return "", nil, fmt.Errorf("method must be provided")
	}

	for _, req := range cfg.Backend {
		if req.Name == "" {
			req.Name = req.ResponseBody
		}

		if req.Name == "" {
			return "", nil, fmt.Errorf("name or response_body must be provided")
		}
	}

	procs := make([]handler.RenderParser, 0, len(cfg.Backend))
	graph := createDepGraph(cfg.Backend)

	backends, err := topSortDFS(cfg.Backend)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sort backends: %w", err)
	}

	for _, procCfg := range backends {
		p, err := processor.New(procCfg)

		if err != nil {
			return "", nil, fmt.Errorf("failed to create processor: %w", err)
		}

		procs = append(procs, p)
	}

	factory := createComposerFactory(graph)

	return cfg.Method, handler.New(valid, procs, factory), nil
}

// topSortDFS performs a topological sort on a slice of BackendConfig using Depth-First Search (DFS).
// It takes a slice of BackendConfig as input and returns a sorted slice of BackendConfig and an error.
// It returns an error if a circular dependency is detected among the BackendConfig elements.
// Each BackendConfig element must have a unique ResponseBody and a list of dependencies specified in DependsOn.
func topSortDFS(be []*processor.Config) ([]*processor.Config, error) {
	graph := createDepGraph(be)
	visited := make(map[string]bool, len(be))
	recStack := make(map[string]bool, len(be))
	sorted := make([]*processor.Config, 0, len(be))

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
func createDepGraph(be []*processor.Config) map[string][]string {
	graph := make(map[string][]string)

	for _, b := range be {
		graph[b.Name] = b.DependsOn
	}

	return graph
}

func createComposerFactory(graph map[string][]string) func(core.Waiter) handler.WaitComposer {
	return func(waiter core.Waiter) handler.WaitComposer {
		return composer.New(graph, waiter)
	}
}
