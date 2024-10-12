package handler

import "html/template"

type Handler struct {
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
