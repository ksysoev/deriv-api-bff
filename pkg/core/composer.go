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
	wg   sync.WaitGroup
	mu   sync.Mutex
	resp map[string]any
	err  error
}

func NewComposer() *Composer {
	return &Composer{
		resp: make(map[string]any),
	}
}

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

func (c *Composer) Response(req_id *int) ([]byte, error) {
	c.wg.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err == nil {
		if req_id != nil {
			c.resp["req_id"] = *req_id
		}

		data, err := json.Marshal(c.resp)
		if err != nil {
			return nil, fmt.Errorf("fail to marshal response: %w", err)
		}

		return data, nil
	}

	var err *APIError
	if errors.As(c.err, &err) {
		return err.Response(req_id)
	}

	return nil, c.err
}

func (c *Composer) setError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return
	}

	c.err = err
}
