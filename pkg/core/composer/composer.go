package composer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

type Composer struct {
	err      error
	rawResps map[string]map[string]any
	req      map[string]chan struct{}
	resp     map[string]any
	wg       sync.WaitGroup
	mu       sync.Mutex
}

// NewComposer creates and returns a new instance of Composer.
// It initializes the Composer with an empty response map.
// It returns a pointer to the newly created Composer instance.
func New() *Composer {
	return &Composer{
		resp: make(map[string]any),
	}
}

// Wait listens for a response on the provided channel and processes it using the given parser.
// It takes a context (ctx) of type context.Context, a parser of type handler.Parser, and a response channel (respChan) of type <-chan []byte.
// It does not return any values but may set an error on the Composer if the context is done or if parsing the response fails.
func (c *Composer) Wait(ctx context.Context, name string, parser handler.Parser, respChan <-chan []byte) {
	c.wg.Add(1)

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
}

func (c *Composer) ComposeDependencies(ctx context.Context, dependsOn []string) (map[string]any, error) {
	c.mu.Lock()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	for _, name := range dependsOn {
		c.wg.Add(1)
		done := c.addRequest(name)

		go func() {
			defer c.wg.Done()

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

	return c.resp, c.err
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
