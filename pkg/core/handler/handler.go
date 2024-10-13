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
	Render(w io.Writer, reqID int64, params map[string]any) error
	Parse(data []byte) (map[string]any, error)
}

type Composer interface {
	WaitResponse(ctx context.Context, parser func([]byte) (map[string]any, error), respChan <-chan []byte)
	Response() (map[string]any, error)
}

type Handler struct {
	validator   Validator
	processors  []Processor
	newComposer func() Composer
}

type request struct {
	id       int64
	respChan <-chan []byte
	parser   func([]byte) (map[string]any, error)
	data     []byte
}

func New(val Validator, proc []Processor, composeFactory func() Composer) *Handler {
	return &Handler{
		validator:   val,
		processors:  proc,
		newComposer: composeFactory,
	}
}

func (h *Handler) Handle(ctx context.Context, params map[string]any, watcher func() (reqID int64, respChan <-chan []byte), send func([]byte) error) (map[string]any, error) {
	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	comp := h.newComposer()

	for req := range h.requests(ctx, params, watcher) {
		comp.WaitResponse(ctx, req.parser, req.respChan)

		if err := send(req.data); err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	return comp.Response()
}

func (h *Handler) requests(ctx context.Context, params map[string]any, watcher func() (reqID int64, respChan <-chan []byte)) iter.Seq[request] {
	var buf bytes.Buffer

	return func(yield func(request) bool) {
		for _, proc := range h.processors {
			if ctx.Err() != nil {
				return
			}
			reqID, respChan := watcher()

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := proc.Render(&buf, reqID, params); err != nil {
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
	}
}