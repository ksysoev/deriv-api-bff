package handler

import (
	"fmt"
	"html/template"
)

type Handler struct {
	validator *validator
	Requests  map[string]*RequestRunConfig
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
		Requests:  cfg.Requets,
	}, nil
}
