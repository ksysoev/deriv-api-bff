package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"sync"

	"github.com/ksysoev/wasabi"
)

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
	ReqID  int
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
	mu        sync.Mutex
}

func (r *RequesIter) Next(id int) ([]byte, chan []byte, error) {
	if r.pos >= len(r.reqs) {
		return nil, nil, io.EOF
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
		return nil, nil, err
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		resp := <-respChan

		if resp == nil {
			slog.Info("Response is nil")
			return
		}

		respBody, err := req.ParseResp(resp)
		if err != nil {
			slog.Info("Fail to parse response", "error", err, "req", body)
			return
		}

		r.mu.Lock()
		defer r.mu.Unlock()

		for _, key := range req.allow {
			if _, ok := respBody[key]; !ok {
				slog.Info("Response body does not contain key", "key", key)
				return
			}

			r.finalResp[key] = respBody[key]
		}
	}()

	return body, respChan, nil
}

func (r *RequesIter) GetResp() map[string]any {
	r.wg.Wait()

	return r.finalResp
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
		fmt.Println("Response data", string(data))
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
