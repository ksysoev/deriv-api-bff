package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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
		reqs:   requests,
		params: r.Params,
	}, nil
}

type RequesIter struct {
	pos    int
	reqs   []*Request
	params map[string]any
}

func (r *RequesIter) Next(id int) ([]byte, error) {
	if r.pos >= len(r.reqs) {
		return nil, nil
	}

	req := r.reqs[r.pos]
	r.pos++

	data := TemplateData{
		Params: r.params,
		ReqID:  id,
	}

	return req.Render(data)
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
