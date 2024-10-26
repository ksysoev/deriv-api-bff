package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/wolfeidau/jsontemplate"
)

type Tmpl struct {
	jt *jsontemplate.Template
}

// New creates a new Tmpl instance from a raw JSON template string.
// It takes tmplRaw of type string, which is the raw JSON template.
// It returns a pointer to Tmpl and an error.
// It returns an error if the template parsing fails.
func New(tmplRaw string) (*Tmpl, error) {
	tmpl, err := jsontemplate.NewTemplate(tmplRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request template: %w", err)
	}

	return &Tmpl{
		jt: tmpl,
	}, nil
}

// Must creates a new Tmpl instance from the provided raw template string.
// It takes tmplRaw of type string, which is the raw template content.
// It returns a pointer to a Tmpl instance.
// It panics if there is an error during the creation of the Tmpl instance.
func Must(tmplRaw string) *Tmpl {
	tmpl, err := New(tmplRaw)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// Execute processes the provided data using the template and returns the result as a byte slice.
// It takes one parameter: data of type []byte, which represents the input data to be processed by the template.
// It returns a byte slice containing the processed data and an error if the template execution fails.
// It returns an error if the template execution encounters any issues, such as invalid data or template errors.
func (t *Tmpl) Execute(params any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	jData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request template: %w", err)
	}

	if _, err := t.jt.Execute(buf, jData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
