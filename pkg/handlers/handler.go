package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"sync"

	"github.com/ksysoev/wasabi"
)

var ErrIterDone = errors.New("iteration done")

type CallHandler struct {
	calls map[string]*CallRunConfig
}

type CallRunConfig struct {
	params   map[string]string
	requests map[string]*RequestRunConfig
}

type RequestRunConfig struct {
	tmplt        *template.Template
	allow        []string
	responseBody string
}

type TemplateData struct {
	Params map[string]any
	ReqID  int64
}

func NewCallHandler(config *HandlersConfig) (*CallHandler, error) {
	h := &CallHandler{
		calls: make(map[string]*CallRunConfig),
	}

	for _, call := range config.Calls {
		c := &CallRunConfig{
			requests: make(map[string]*RequestRunConfig),
		}
		for _, req := range call.Backend {
			tmplt, err := template.New("request").Parse(req.RequestTemplate)
			if err != nil {
				return nil, err
			}
			c.requests[req.ResponseBody] = &RequestRunConfig{
				tmplt:        tmplt,
				allow:        req.Allow,
				responseBody: req.ResponseBody,
			}
		}
		h.calls[call.Method] = c
	}

	return h, nil
}

func (h *CallHandler) Process(req wasabi.Request) (*RequesIter, error) {
	method := req.RoutingKey()

	call, ok := h.calls[method]

	if !ok {
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	data := req.Data()

	type D struct {
		Params map[string]interface{} `json:"params"`
	}
	var r D

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	requests := make([]*Request, 0, len(call.requests))
	for _, req := range call.requests {
		requests = append(requests, &Request{
			tempate:      req.tmplt,
			allow:        req.allow,
			responseBody: req.responseBody,
		})
	}

	return &RequesIter{
		reqs:      requests,
		params:    r.Params,
		finalResp: make(map[string]any),
	}, nil
}

type RequesIter struct {
	wg        sync.WaitGroup
	pos       int
	reqs      []*Request
	params    map[string]any
	finalResp map[string]any
	err       error
	mu        sync.Mutex
}

func (r *RequesIter) Next(ctx context.Context, id int64) ([]byte, chan []byte, error) {
	if r.pos >= len(r.reqs) {
		return nil, nil, ErrIterDone
	}

	req := r.reqs[r.pos]
	r.pos++

	data := TemplateData{
		Params: r.params,
		ReqID:  id,
	}

	respChan := make(chan []byte, 1)

	body, err := req.Render(data)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to render request %s: %w", req.responseBody, err)
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		select {
		case <-ctx.Done():
			r.mu.Lock()
			defer r.mu.Unlock()

			r.err = ctx.Err()

			return
		case resp := <-respChan:
			respBody, err := req.ParseResp(resp)

			r.mu.Lock()
			defer r.mu.Unlock()

			if err != nil {
				r.err = fmt.Errorf("fail to parse response %s: %w", req.responseBody, err)

				return
			}

			for _, key := range req.allow {
				if _, ok := respBody[key]; !ok {
					slog.Warn("Response body does not contain key", "key", key)
					return
				}

				r.finalResp[key] = respBody[key]
			}
		}
	}()

	return body, respChan, nil
}

func (r *RequesIter) WaitResp() (map[string]any, error) {
	r.wg.Wait()

	return r.finalResp, r.err
}

type Request struct {
	tempate      *template.Template
	allow        []string
	responseBody string
	mu           sync.Mutex
}

func (r *Request) Render(data TemplateData) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var buf bytes.Buffer
	err := r.tempate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Request) ParseResp(data []byte) (map[string]any, error) {
	var rdata map[string]any

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return nil, err
	}

	rb, ok := rdata[r.responseBody]
	if !ok {
		for _, v := range rdata {
			slog.Info("Response body", "key", v)
		}
		return nil, fmt.Errorf("response body not found")
	}
	respBody, ok := rb.(map[string]any)

	if !ok {
		return nil, fmt.Errorf("response body is not an object")
	}

	return respBody, nil

}
