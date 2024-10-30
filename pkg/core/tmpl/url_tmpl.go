package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/valyala/fasttemplate"
	"github.com/wolfeidau/jsontemplate"
)

type URLTmpl struct {
	tmpl *fasttemplate.Template
}

func NewURLTmpl(tmpl string) (*URLTmpl, error) {
	t, err := fasttemplate.NewTemplate(tmpl, "${", "}")
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &URLTmpl{tmpl: t}, nil
}

func (t *URLTmpl) Execute(params any) ([]byte, error) {
	jData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request template: %w", err)
	}

	doc := jsontemplate.NewDocument(jData)

	buf := bytes.NewBuffer(nil)

	_, err = t.tmpl.ExecuteFunc(buf, func(w io.Writer, path string) (int, error) {
		v, err := doc.Read(path + ";escape")
		if err != nil {
			return 0, err
		}

		str, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("expected string, got %T", v)
		}

		return w.Write([]byte(str))
	})

	return buf.Bytes(), err
}
