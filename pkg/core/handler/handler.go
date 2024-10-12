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

type Handler struct {
	validator  Validator
	processors []Processor
}

type TemplateData struct {
	Params map[string]any
	ReqID  int64
}

func NewHandler(val Validator, proc []Processor) *Handler {
	return &Handler{
		validator:  val,
		processors: proc,
	}
}

func (h *Handler) Requests(ctx context.Context, params map[string]any, getID func() int64) (iter.Seq[[]byte], error) {
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
