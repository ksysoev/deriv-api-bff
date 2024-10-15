package composer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handler"
)

type Composer struct {
	err  error
	resp map[string]any
	wg   sync.WaitGroup
	mu   sync.Mutex
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
func (c *Composer) Wait(ctx context.Context, parser handler.Parser, respChan <-chan []byte) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		select {
		case <-ctx.Done():
			c.setError(ctx.Err())
		case resp := <-respChan:
			data, err := parser(resp)
			if err != nil {
				c.setError(fmt.Errorf("fail to parse response: %w", err))
				return
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			for key, value := range data {
				if _, ok := c.resp[key]; ok {
					slog.Warn("duplicate key", slog.String("key", key))
				}

				c.resp[key] = value
			}
		}
	}()
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
func (c *Composer) setError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return
	}

	c.err = err
}
