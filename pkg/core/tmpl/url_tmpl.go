package tmpl

import (
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

// Execute generates a URL string by executing a template with the provided parameters.
// It takes params of type any, which are marshaled into JSON and used to fill the template.
// It returns a string containing the generated URL and an error if the operation fails.
// It returns an error if the parameters cannot be marshaled into JSON, if the template execution fails,
// or if the template path does not resolve to a string value.
func (t *URLTmpl) Execute(params any) (string, error) {
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

		str, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("expected string, got %T", v)
		}

		escapedStr := url.QueryEscape(str)

		return w.Write([]byte(escapedStr))
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return str, nil
}
