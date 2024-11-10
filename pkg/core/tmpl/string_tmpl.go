package tmpl

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/valyala/fasttemplate"
	"github.com/wolfeidau/jsontemplate"
)

type StrTmpl struct {
	tmpl *fasttemplate.Template
}

func NewStrTmpl(tmpl string) (*StrTmpl, error) {
	t, err := fasttemplate.NewTemplate(tmpl, "${", "}")
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &StrTmpl{tmpl: t}, nil
}

func MustNewStringTmpl(tmplRaw string) *StrTmpl {
	tmpl, err := NewStrTmpl(tmplRaw)
	if err != nil {
		panic(err)
	}

	return tmpl
}

func (t *StrTmpl) Execute(params any) (string, error) {
	jData, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request template: %w", err)
	}

	doc := jsontemplate.NewDocument(jData)

	str, err := t.tmpl.ExecuteFuncStringWithErr(func(w io.Writer, path string) (int, error) {
		v, err := doc.Read(path + ";escape")
		if err != nil {
			return 0, err
		}

		if str, ok := v.(string); ok {
			return w.Write([]byte(str))
		}

		return 0, fmt.Errorf("expected string, got %T", v)
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return str, nil
}
