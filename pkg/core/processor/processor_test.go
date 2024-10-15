package processor

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	cfg := &Config{
		Tmplt:        tmpl,
		FieldMap:     map[string]string{"key1": "mappedKey1"},
		ResponseBody: "data",
		Allow:        []string{"key1", "key2"},
	}

	processor := New(cfg)

	assert.NotNil(t, processor)
	assert.Equal(t, tmpl, processor.tmpl)
	assert.Equal(t, cfg.FieldMap, processor.fieldMap)
	assert.Equal(t, cfg.ResponseBody, processor.responseBody)
	assert.Equal(t, cfg.Allow, processor.allow)
}

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
		expected     map[string]any
		name         string
		responseBody string
		jsonData     string
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
func TestProcessor_Parse_Success(t *testing.T) {
	tests := []struct {
		fieldMap     map[string]string
		expected     map[string]any
		name         string
		responseBody string
		jsonData     string
		allow        []string
	}{
		{
			name:         "allowed fields with field mapping",
			responseBody: "data",
			fieldMap:     map[string]string{"key1": "mappedKey1"},
			allow:        []string{"key1", "key2"},
			jsonData:     `{"data": {"key1": "value1", "key2": "value2"}}`,
			expected:     map[string]any{"mappedKey1": "value1", "key2": "value2"},
		},
		{
			name:         "allowed fields without field mapping",
			responseBody: "data",
			fieldMap:     nil,
			allow:        []string{"key1", "key2"},
			jsonData:     `{"data": {"key1": "value1", "key2": "value2"}}`,
			expected:     map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:         "missing allowed fields",
			responseBody: "data",
			fieldMap:     nil,
			allow:        []string{"key3"},
			jsonData:     `{"data": {"key1": "value1", "key2": "value2"}}`,
			expected:     map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &Processor{
				responseBody: tt.responseBody,
				fieldMap:     tt.fieldMap,
				allow:        tt.allow,
			}

			result, err := rp.Parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_Parse_Error(t *testing.T) {
	rp := &Processor{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.Parse([]byte(jsonData))
	assert.Error(t, err)
}
