package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/valyala/fasttemplate"
	"github.com/wolfeidau/jsontemplate"
)

type URLTmpl struct {
	tmpl *fasttemplate.Template
}

// NewURLTmpl creates a new URLTmpl instance by parsing the provided template string.
// It takes tmpl of type string, which represents the template to be parsed.
// It returns a pointer to a URLTmpl and an error.
// It returns an error if the template parsing fails.
func NewURLTmpl(tmpl string) (*URLTmpl, error) {
	t, err := fasttemplate.NewTemplate(tmpl, "${", "}")
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &URLTmpl{tmpl: t}, nil
}

// Execute processes the given parameters using a JSON template and returns the resulting byte slice.
// It takes params of type any, which are marshaled into JSON format.
// It returns a byte slice containing the processed template and an error if any occurs during processing.
// It returns an error if the parameters cannot be marshaled into JSON, if the template execution fails, or if the expected string type is not met during template processing.
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

		escapedStr := url.QueryEscape(str)

		return w.Write([]byte(escapedStr))
	})

	return buf.Bytes(), err
}
