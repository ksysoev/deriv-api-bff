package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
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
func NewComposer() *Composer {
	return &Composer{
		resp: make(map[string]any),
	}
}

// WaitResponse waits for a response on the provided channel and processes it.
// It takes a context (ctx) of type context.Context, a request processor (req) of type *RequestProcessor, and a response channel (respChan) of type <-chan []byte.
// It does not return any values.
// It sets an error if the context is done before a response is received or if there is a failure in parsing the response.
// If the response body does not contain expected keys, it logs a warning and continues processing other keys.
func (c *Composer) WaitResponse(ctx context.Context, req *RequestProcessor, respChan <-chan []byte) {
	c.wg.Add(1)
	defer c.wg.Done()

	select {
	case <-ctx.Done():
		c.setError(ctx.Err())
	case resp := <-respChan:
		respBody, err := req.ParseResp(resp)

		if err != nil {
			c.setError(fmt.Errorf("fail to parse response %s: %w", req.responseBody, err))

			return
		}

		for _, key := range req.allow {
			if _, ok := respBody[key]; !ok {
				slog.Warn("Response body does not contain expeted key", slog.String("key", key), slog.String("response_body", req.responseBody))
				continue
			}

			destKey := key

			if req.fieldsMap != nil {
				if mappedKey, ok := req.fieldsMap[key]; ok {
					destKey = mappedKey
				}
			}

			c.resp[destKey] = respBody[key]
		}
	}
}

// Response generates a JSON response based on the Composer's state.
// It takes req_id of type *int, which is an optional request identifier.
// It returns a byte slice containing the JSON response and an error if any occurs.
// It returns an error if the response cannot be marshaled into JSON or if there is an existing error in the Composer.
// If the Composer has an APIError, it delegates the response generation to the APIError's Response method.
func (c *Composer) Response(reqID *int) ([]byte, error) {
	c.wg.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err == nil {
		if reqID != nil {
			c.resp["req_id"] = *reqID
		}

		data, err := json.Marshal(c.resp)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal response: %w", err)
		}

		return data, nil
	}

	var err *APIError
	if errors.As(c.err, &err) {
		return err.Response(reqID)
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
