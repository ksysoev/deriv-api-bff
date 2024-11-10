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

// NewStrTmpl creates a new StrTmpl instance by parsing the provided template string.
// It takes tmpl of type string, which is the template to be parsed.
// It returns a pointer to StrTmpl and an error.
// It returns an error if the template parsing fails.
func NewStrTmpl(tmpl string) (*StrTmpl, error) {
	t, err := fasttemplate.NewTemplate(tmpl, "${", "}")
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &StrTmpl{tmpl: t}, nil
}

// MustNewStrTmpl creates a new StrTmpl from the provided raw template string.
// It takes tmplRaw of type string.
// It returns a pointer to StrTmpl.
// It panics if there is an error during the creation of the StrTmpl.
func MustNewStrTmpl(tmplRaw string) *StrTmpl {
	tmpl, err := NewStrTmpl(tmplRaw)
	if err != nil {
		panic(err)
	}

	return tmpl
}

// Execute processes the given parameters using a JSON template and returns the resulting string.
// It takes params of type any, which represents the data to be used in the template.
// It returns a string containing the processed template and an error if the operation fails.
// It returns an error if the parameters cannot be marshaled into JSON, if the template execution fails,
// or if the template path does not resolve to a string.
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
