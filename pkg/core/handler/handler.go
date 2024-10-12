package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
)

type Validator interface {
	Validate(data map[string]any) error
}

type Processor interface {
	Render(w io.Writer, data *TemplateData) error
	Parse(data []byte) (map[string]any, error)
}

type Composer interface {
	WaitResponse(ctx context.Context, parser func([]byte) (map[string]any, error), respChan <-chan []byte)
	Response() (map[string]any, error)
}

type ResponseWatcher func() (reqID int64, respChan <-chan []byte)
type Sender func([]byte) error

type Handler struct {
	validator   Validator
	processors  []Processor
	newComposer func() Composer
}

type TemplateData struct {
	Params map[string]any
	ReqID  int64
}

type request struct {
	id       int64
	respChan <-chan []byte
	parser   func([]byte) (map[string]any, error)
	data     []byte
}

func NewHandler(val Validator, proc []Processor, composeFactory func() Composer) *Handler {
	return &Handler{
		validator:   val,
		processors:  proc,
		newComposer: composeFactory,
	}
}

func (h *Handler) Handle(ctx context.Context, params map[string]any, watcher ResponseWatcher, send Sender) (map[string]any, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	iter, err := h.requests(ctx, params, watcher)
	if err != nil {
		return nil, fmt.Errorf("failed to create request iterator: %w", err)
	}

	comp := h.newComposer()

	for req := range iter {
		comp.WaitResponse(ctx, req.parser, req.respChan)

		if err := send(req.data); err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	return comp.Response()
}

func (h *Handler) requests(ctx context.Context, params map[string]any, watcher ResponseWatcher) (iter.Seq[request], error) {
	var buf bytes.Buffer

	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	return func(yield func(request) bool) {
		for _, proc := range h.processors {
			if ctx.Err() != nil {
				return
			}
			reqID, respChan := watcher()

			d := TemplateData{
				Params: params,
				ReqID:  reqID,
			}

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := proc.Render(&buf, &d); err != nil {
				// TODO: add prevalidating template on startup to avoid this error in runtime
				panic(fmt.Sprintf("template execution failed: %v", err))
			}

			request := request{
				id:       reqID,
				respChan: respChan,
				parser:   proc.Parse,
				data:     buf.Bytes(),
			}

			yield(request)
		}
	}, nil
}
