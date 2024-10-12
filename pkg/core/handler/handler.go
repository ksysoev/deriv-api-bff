package handler

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"iter"
)

type Handler struct {
	validator *validator
	reqs      map[string]*RequestRunConfig
}

type HandlerConfig struct {
	Params  ValidatorConfig
	Requets map[string]*RequestRunConfig
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

func NewHandler(cfg *HandlerConfig) (*Handler, error) {
	v, err := NewValidator(&cfg.Params)
	if err != nil {
		return nil, err
	}

	if len(cfg.Requets) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}

	return &Handler{
		validator: v,
		reqs:      cfg.Requets,
	}, nil
}

func (h *Handler) Requests(ctx context.Context, params map[string]any, getID func() int64) (iter.Seq[[]byte], error) {
	var buf bytes.Buffer

	if err := h.validator.Validate(params); err != nil {
		return nil, err
	}

	return func(yield func([]byte) bool) {
		for _, req := range h.reqs {
			if ctx.Err() != nil {
				return
			}

			d := TemplateData{
				Params: params,
				ReqID:  getID(),
			}

			// TODO: check for race conditions here that iterator blocks until the previous request is sent
			buf.Reset()

			if err := req.Tmplt.Execute(&buf, d); err != nil {
				// TODO: add prevalidating template on startup to avoid this error in runtime
				panic(fmt.Sprintf("template execution failed: %v", err))
			}

			yield(buf.Bytes())
		}
	}, nil
}
