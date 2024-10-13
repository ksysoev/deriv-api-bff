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

// WaitResponse waits for a response on the provided channel and processes it.
// It takes a context (ctx) of type context.Context, a request processor (req) of type *RequestProcessor, and a response channel (respChan) of type <-chan []byte.
// It does not return any values.
// It sets an error if the context is done before a response is received or if there is a failure in parsing the response.
// If the response body does not contain expected keys, it logs a warning and continues processing other keys.
func (c *Composer) Wait(ctx context.Context, parser handler.Parser, respChan <-chan []byte) {
	c.wg.Add(1)
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
}

// Response generates a JSON response based on the Composer's state.
// It takes req_id of type *int, which is an optional request identifier.
// It returns a byte slice containing the JSON response and an error if any occurs.
// It returns an error if the response cannot be marshaled into JSON or if there is an existing error in the Composer.
// If the Composer has an APIError, it delegates the response generation to the APIError's Response method.
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
