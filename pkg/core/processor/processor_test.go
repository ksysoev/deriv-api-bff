package processor

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor_Render(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	rp := &Processor{
		tmpl: tmpl,
	}

	params := map[string]any{"key1": "value1", "key2": "value2"}
	reqID := 12345
	expected := "Params: map[key1:value1 key2:value2], ReqID: 12345"

	var buf bytes.Buffer

	err = rp.Render(&buf, int64(reqID), params)
	assert.NoError(t, err)
	assert.Equal(t, expected, buf.String())
}

func TestProcessor_parse_Success(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		jsonData     string
		expected     map[string]any
	}{
		{
			name:         "object",
			responseBody: "data",
			jsonData:     `{"data": {"key1": "value1", "key2": "value2"}}`,
			expected:     map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:         "array",
			responseBody: "data",
			jsonData:     `{"data": [{"key1": "value1"}, {"key2": "value2"}]}`,
			expected:     map[string]any{"list": []any{map[string]any{"key1": "value1"}, map[string]any{"key2": "value2"}}},
		},
		{
			name:         "scalar",
			responseBody: "data",
			jsonData:     `{"data": "value"}`,
			expected:     map[string]any{"value": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &Processor{
				responseBody: tt.responseBody,
			}

			result, err := rp.parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_parse_Error(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_ResponseBodyNotFound(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"key": "value"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_UnexpectedFormat(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{invalid json}`

	result, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
	assert.Nil(t, result)
}
