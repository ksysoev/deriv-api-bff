package core

import (
	"context"
	"errors"
	"fmt"
	"html/template"
)

type CallsRepo interface {
	GetCall(method string) *CallRunConfig
}

var ErrIterDone = errors.New("iteration done")

type CallHandler struct {
	repo CallsRepo
}

type CallRunConfig struct {
	Requests map[string]*RequestRunConfig
}

type RequestRunConfig struct {
	Tmplt        *template.Template
	FieldMap     map[string]string
	ResponseBody string
	Allow        []string
}

type TemplateData struct {
	Params map[string]any
	ReqID  int64
}

// NewCallHandler initializes a new CallHandler based on the provided configuration.
// It takes a config parameter of type *Config which contains the necessary setup for calls and requests.
// It returns a pointer to CallHandler and an error if there is an issue parsing the request templates.
// It returns an error if any request template in the configuration cannot be parsed.
func NewCallHandler(repo CallsRepo) *CallHandler {
	return &CallHandler{
		repo: repo,
	}
}

// Process handles the processing of a request based on its routing key.
// It takes a req of type *Request and returns a *RequesIter and an error.
// It returns an error if the method specified in the request's routing key is unsupported.
// The function initializes a context with cancellation and prepares a list of request processors.
func (h *CallHandler) Process(req *Request) (*RequesIter, error) {
	method := req.RoutingKey()

	call := h.repo.GetCall(method)
	if call == nil {
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	requests := make([]*RequestProcessor, 0, len(call.Requests))
	for _, req := range call.Requests {
		requests = append(requests, &RequestProcessor{
			tempate:      req.Tmplt,
			allow:        req.Allow,
			fieldsMap:    req.FieldMap,
			responseBody: req.ResponseBody,
		})
	}

	ctx, cancel := context.WithCancel(req.Context())

	return &RequesIter{
		ctx:       ctx,
		cancel:    cancel,
		reqs:      requests,
		params:    req.Params,
		finalResp: make(map[string]any),
		composer:  NewComposer(),
	}, nil
}

type RequesIter struct {
	ctx       context.Context
	cancel    context.CancelFunc
	params    map[string]any
	finalResp map[string]any
	composer  *Composer
	reqs      []*RequestProcessor
	pos       int
}

// HasNext checks if there are more requests to process in the RequesIter.
// It returns a boolean value: true if there are more requests, false otherwise.
func (r *RequesIter) HasNext() bool {
	return r.pos < len(r.reqs)
}

// Next retrieves the next request in the iterator, renders it with the provided data, and sends it to the response channel.
// It takes an id of type int64 and a respChan of type <-chan []byte.
// It returns a byte slice containing the rendered request body and an error if rendering fails or if the iterator is done.
// It returns ErrIterDone if there are no more requests in the iterator.
func (r *RequesIter) Next(id int64, respChan <-chan []byte) ([]byte, error) {
	if r.pos >= len(r.reqs) {
		return nil, ErrIterDone
	}

	req := r.reqs[r.pos]
	r.pos++

	data := TemplateData{
		Params: r.params,
		ReqID:  id,
	}

	body, err := req.Render(data)
	if err != nil {
		return nil, fmt.Errorf("fail to render request %s: %w", req.responseBody, err)
	}

	go r.composer.WaitResponse(r.ctx, req, respChan)

	return body, nil
}

// WaitResp waits for a response corresponding to the given request ID.
// It takes req_id of type *int, which is a pointer to the request ID.
// It returns a byte slice containing the response data and an error if the response cannot be retrieved.
// It returns an error if the request ID is invalid or if there is a failure in the response retrieval process.
func (r *RequesIter) WaitResp(reqID *int) ([]byte, error) {
	return r.composer.Response(reqID)
}
