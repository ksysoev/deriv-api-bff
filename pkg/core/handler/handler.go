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
	DependsOn() []string
	Render(w io.Writer, reqID int64, params map[string]any, deps map[string]any) error
	Parse(data []byte) (map[string]any, map[string]any, error)
}

type WaitComposer interface {
	Wait(ctx context.Context, name string, parser Parser, respChan <-chan []byte)
	ComposeDependencies(ctx context.Context, dependsOn []string) (map[string]any, error)
	Compose() (map[string]any, error)
}

type Handler struct {
	validator   Validator
	newComposer func() WaitComposer
	processors  []RenderParser
}

type request struct {
	name     string
	respChan <-chan []byte
	parser   func([]byte) (map[string]any, map[string]any, error)
	data     []byte
	id       int64
}

// New creates a new instance of Handler with the provided validator, processors, and composer factory function.
// It takes three parameters: val of type Validator, proc which is a slice of RenderParser, and composeFactory which is a function returning a WaitComposer.
// It returns a pointer to a Handler initialized with the provided parameters.
func New(val Validator, proc []RenderParser, composeFactory func() WaitComposer) *Handler {
	return &Handler{
		validator:   val,
		processors:  proc,
		newComposer: composeFactory,
	}
}

// Handle processes incoming requests, validates them, and sends them to the appropriate handler.
// It takes ctx of type context.Context, params of type map[string]any, watcher of type core.Waiter, and send of type core.Sender.
// It returns a map[string]any containing the composed response and an error if any occurs during validation or sending requests.
// It returns an error if the validation of params fails or if sending a request fails.
func (h *Handler) Handle(ctx context.Context, params map[string]any, watcher core.Waiter, send core.Sender) (map[string]any, error) {
	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	comp := h.newComposer()

	for req := range h.requests(ctx, params, watcher, comp) {
		comp.Wait(ctx, req.name, req.parser, req.respChan)

		if err := send(ctx, req.data); err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	return comp.Compose()
}

// requests generates a sequence of requests based on the provided context, parameters, and watcher.
// It takes ctx of type context.Context, params of type map[string]any, and watcher of type core.Waiter.
// It returns an iterator function that yields requests of type request.
// It panics if there is an error in rendering the processor template.
// Special behavior includes checking for context cancellation and resetting the buffer for each processor.
func (h *Handler) requests(ctx context.Context, params map[string]any, watcher core.Waiter, comp WaitComposer) iter.Seq[request] {
	var buf bytes.Buffer

	return func(yield func(request) bool) {
		for _, proc := range h.processors {
			if ctx.Err() != nil {
				return
			}

			depsOn := proc.DependsOn()

			depResuls := make(map[string]any)
			if len(depsOn) > 0 {
				var err error
				if depResuls, err = comp.ComposeDependencies(ctx, depsOn); err != nil {
					return
				}
			}

			reqID, respChan := watcher()

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := proc.Render(&buf, reqID, params, depResuls); err != nil {
				// TODO: add prevalidating template on startup to avoid this error in runtime
				panic(fmt.Sprintf("template execution failed: %v", err))
			}

			request := request{
				name:     proc.Name(),
				id:       reqID,
				respChan: respChan,
				parser:   proc.Parse,
				data:     buf.Bytes(),
			}

			if !yield(request) {
				return
			}
		}
	}
}
