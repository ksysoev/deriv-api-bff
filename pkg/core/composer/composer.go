package composer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

type Composer struct {
	depGraph map[string][]string
	err      error
	rawResps map[string]any
	req      map[string]chan struct{}
	resp     map[string]any
	waiter   core.Waiter
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// New creates and returns a new instance of Composer.
// It takes depGraph of type map[string][]string which represents the dependency graph,
// and waiter of type core.Waiter which is used to manage synchronization.
// It returns a pointer to a Composer struct initialized with the provided depGraph and waiter.
func New(depGraph map[string][]string, waiter core.Waiter) *Composer {
	return &Composer{
		depGraph: depGraph,
		waiter:   waiter,
		resp:     make(map[string]any),
		req:      make(map[string]chan struct{}),
		rawResps: make(map[string]any),
	}
}

// Prepare initializes and prepares a request by composing dependencies and parsing the response.
// It takes a context.Context, a string name, and a handler.Parser.
// It returns an int64 request ID, a map of dependency results, and an error if the dependencies cannot be composed or the response cannot be parsed.
// It returns an error if the dependencies cannot be composed or the response cannot be parsed.
func (c *Composer) Prepare(ctx context.Context, name string, parser handler.Parser) (reqID int64, depsResults map[string]any, err error) {
	c.wg.Add(1)

	depsResults, err = c.composeDependencies(ctx, name)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to compose dependencies: %w", err)
	}

	reqID, respChan := c.waiter()

	go func() {
		defer c.wg.Done()
		select {
		case <-ctx.Done():
			c.setError(name, ctx.Err())
		case resp := <-respChan:
			rawResp, data, err := parser(resp)
			if err != nil {
				c.setError(name, fmt.Errorf("fail to parse response: %w", err))
				return
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			c.rawResps[name] = rawResp

			for key, value := range data {
				if _, ok := c.resp[key]; ok {
					//TODO: Move fields uniqueness check to the config validation step
					slog.Warn("duplicate key", slog.String("key", key))
				}

				c.resp[key] = value
			}

			c.doneRequest(name)
		}
	}()

	return reqID, depsResults, nil
}

// composeDependencies composes the dependencies for a given name by executing
// the required dependency functions concurrently and collecting their results.
// It takes a context.Context and a string name as parameters.
// It returns a map[string]any containing the results of the dependencies and an error if any dependency fails.
// It returns an error if any of the dependencies encounter an error during execution.
// The function locks the mutex to ensure thread safety and uses a wait group to wait for all goroutines to complete.
func (c *Composer) composeDependencies(ctx context.Context, name string) (map[string]any, error) {
	dependsOn := c.depGraph[name]

	if len(dependsOn) == 0 {
		return make(map[string]any), nil
	}

	c.mu.Lock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	for _, name := range dependsOn {
		wg.Add(1)

		done := c.addRequest(name)

		go func() {
			defer wg.Done()

			select {
			case <-ctx.Done():
			case <-done:
				c.mu.Lock()
				err := c.err
				c.mu.Unlock()

				if err != nil {
					cancel()
				}
			}
		}()
	}
	c.mu.Unlock()

	wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.rawResps, c.err
}

// Compose waits for all requests to finish and then returns the composed result.
// It does not take any parameters.
// It returns a map[string]any containing the composed result and an error if any occurred during composition.
// It returns an error if there was an issue during the composition process.
func (c *Composer) Compose() (map[string]any, error) {
	c.wg.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err == nil {
		return c.resp, nil
	}

	return nil, c.err
}

// setError sets the error for the Composer instance if it is not already set.
// It takes one parameter: err of type error.
// It does not return any values.
// If the Composer instance already has an error set, it does nothing.
func (c *Composer) setError(name string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return
	}

	c.err = err

	c.doneRequest(name)
}

func (c *Composer) addRequest(name string) <-chan struct{} {
	if _, ok := c.req[name]; ok {
		return c.req[name]
	}

	c.req[name] = make(chan struct{})

	return c.req[name]
}

func (c *Composer) doneRequest(name string) {
	if ch, ok := c.req[name]; ok {
		close(ch)
		return
	}

	ch := make(chan struct{})
	close(ch)

	c.req[name] = ch
}
