package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type Parser func([]byte) (map[string]any, map[string]any, error)

type Validator interface {
	Validate(data map[string]any) error
}

type RenderParser interface {
	Name() string
	Render(w io.Writer, reqID int64, params map[string]any, deps map[string]any) error
	Parse(data []byte) (map[string]any, map[string]any, error)
}

type WaitComposer interface {
	Prepare(context.Context, string, Parser) (int64, map[string]any, error)
	Compose() (map[string]any, error)
}

type Handler struct {
	validator   Validator
	newComposer func(core.Waiter) WaitComposer
	processors  []RenderParser
}

type request struct {
	parser func([]byte) (map[string]any, map[string]any, error)
	data   []byte
	id     int64
}

// New creates a new instance of Handler.
// It takes three parameters: val of type Validator, proc which is a slice of RenderParser, and composeFactory which is a function that takes a core.Waiter and returns a WaitComposer.
// It returns a pointer to a Handler.
func New(val Validator, proc []RenderParser, composeFactory func(core.Waiter) WaitComposer) *Handler {
	return &Handler{
		validator:   val,
		processors:  proc,
		newComposer: composeFactory,
	}
}

// Handle processes incoming requests and sends them using the provided sender.
// It takes a context.Context, a map of parameters, a core.Waiter, and a core.Sender.
// It returns a map containing the composed results and an error if any occurs during validation or sending requests.
// It returns an error if the validation of parameters fails or if sending a request fails.
func (h *Handler) Handle(ctx context.Context, params map[string]any, waiter core.Waiter, send core.Sender) (map[string]any, error) {
	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	comp := h.newComposer(waiter)

	for req := range h.requests(ctx, params, comp) {
		if err := send(ctx, req.data); err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	return comp.Compose()
}

// requests generates a sequence of requests based on the provided processors.
// It takes a context `ctx` for managing request lifecycle, a map `params` containing parameters for the requests, and a `comp` of type WaitComposer for preparing the requests.
// It returns an iterator function that yields requests of type `request`.
// The function handles context cancellation and prepares requests using the provided processors. It panics if template execution fails during request rendering.
func (h *Handler) requests(ctx context.Context, params map[string]any, comp WaitComposer) iter.Seq[request] {
	var buf bytes.Buffer

	return func(yield func(request) bool) {
		for _, proc := range h.processors {
			if ctx.Err() != nil {
				return
			}

			reqID, depResuls, err := comp.Prepare(ctx, proc.Name(), proc.Parse)
			if err != nil {
				return
			}

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := proc.Render(&buf, reqID, params, depResuls); err != nil {
				// TODO: add prevalidating template on startup to avoid this error in runtime
				panic(fmt.Sprintf("template execution failed: %v", err))
			}

			request := request{
				id:     reqID,
				parser: proc.Parse,
				data:   buf.Bytes(),
			}

			if !yield(request) {
				return
			}
		}
	}
}
