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
	Response(reqID *int) ([]byte, error)
}

type ResponseWatcher func () reqID int64, respChan <-chan []byte

type Handler struct {
	validator   Validator
	processors  []Processor
	newComposer func() Composer
}

type TemplateData struct {
	Params map[string]any
	ReqID  int64
}

func NewHandler(val Validator, proc []Processor, composeFactory func() Composer) *Handler {
	return &Handler{
		validator:   val,
		processors:  proc,
		newComposer: composeFactory,
	}
}


func (h *Handler) Handle(ctx context.Context, params map[string]any) (map[string]any, error) {
	composer := h.newComposer()
	respChan := make(chan []byte, 1)

	reqs, err := h.requests(ctx, params, composer)
	if err != nil {
		return nil, err
	}

	go composer.WaitResponse(ctx, h.parse, respChan)

	for reqs.HasNext() {
		req, err := reqs.Next()
		if err != nil {
			return nil, err
		}

		respChan <- req
	}

	resp, err := composer.Response(nil)
	if err != nil {
		return nil, err
	}

	return h.parse(resp)

}


func (h *Handler) requests(ctx context.Context, params map[string]any, getID func() int64) (iter.Seq[[]byte], error) {
	var buf bytes.Buffer

	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	return func(yield func([]byte) bool) {
		for _, proc := range h.processors {
			if ctx.Err() != nil {
				return
			}

			d := TemplateData{
				Params: params,
				ReqID:  getID(),
			}

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := proc.Render(&buf, &d); err != nil {
				// TODO: add prevalidating template on startup to avoid this error in runtime
				panic(fmt.Sprintf("template execution failed: %v", err))
			}

			yield(buf.Bytes())
		}
	}, nil
}
