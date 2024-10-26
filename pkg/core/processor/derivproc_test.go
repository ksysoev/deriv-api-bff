package processor

import (
	"context"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeriv(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	assert.NoError(t, err)

	cfg := &Config{
		Tmplt:        tmpl,
		FieldMap:     map[string]string{"key1": "mappedKey1"},
		ResponseBody: "data",
		Allow:        []string{"key1", "key2"},
	}

	processor := NewDeriv(cfg)

	assert.NotNil(t, processor)
	assert.Equal(t, tmpl, processor.tmpl)
	assert.Equal(t, cfg.FieldMap, processor.fieldMap)
	assert.Equal(t, cfg.ResponseBody, processor.responseBody)
	assert.Equal(t, cfg.Allow, processor.allow)
}

func TestProcessor_Render(t *testing.T) {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}, Resp: {{.Resp}}")
	assert.NoError(t, err)

	tmpl1, err := template.New("test").Parse("Params: {{index .Params \"key1\" \"key2\"}}")
	assert.NoError(t, err)

	tests := []struct {
		params   map[string]any
		deps     map[string]any
		tmpl     *template.Template
		name     string
		expected string
		reqID    int64
		wantErr  bool
	}{
		{
			name:     "with params and deps",
			params:   map[string]any{"key1": "value1"},
			deps:     map[string]any{"dep1": "value1"},
			reqID:    12345,
			expected: "Params: map[key1:value1], ReqID: 12345, Resp: map[dep1:value1]",
			wantErr:  false,
		},
		{
			name:     "with nil params and deps",
			params:   nil,
			deps:     nil,
			reqID:    12345,
			expected: "Params: map[], ReqID: 12345, Resp: map[]",
			wantErr:  false,
		},
		{
			name:     "with empty params and deps",
			params:   map[string]any{},
			deps:     map[string]any{},
			reqID:    12345,
			expected: "Params: map[], ReqID: 12345, Resp: map[]",
			wantErr:  false,
		},
		{
			name:     "with only params",
			params:   map[string]any{"key1": "value1"},
			deps:     nil,
			reqID:    12345,
			expected: "Params: map[key1:value1], ReqID: 12345, Resp: map[]",
			wantErr:  false,
		},
		{
			name:     "with only deps",
			params:   nil,
			deps:     map[string]any{"dep1": "value1"},
			reqID:    12345,
			expected: "Params: map[], ReqID: 12345, Resp: map[dep1:value1]",
			wantErr:  false,
		},
		{
			name:     "with error",
			params:   nil,
			deps:     nil,
			reqID:    12345,
			tmpl:     tmpl1,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{
				tmpl: tmpl,
			}

			if tt.tmpl != nil {
				rp.tmpl = tt.tmpl
			}

			ctx := context.Background()
			req, err := rp.Render(ctx, tt.reqID, tt.params, tt.deps)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(req.Data()))
		})
	}
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
			rp := &DerivProc{
				responseBody: tt.responseBody,
			}

			result, err := rp.parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_parse_Error(t *testing.T) {
	rp := &DerivProc{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_ResponseBodyNotFound(t *testing.T) {
	rp := &DerivProc{
		responseBody: "data",
	}

	jsonData := `{"key": "value"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_UnexpectedFormat(t *testing.T) {
	rp := &DerivProc{
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
			rp := &DerivProc{
				responseBody: tt.responseBody,
				fieldMap:     tt.fieldMap,
				allow:        tt.allow,
			}

			_, result, err := rp.Parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_Parse_Error(t *testing.T) {
	rp := &DerivProc{
		responseBody: "data",
	}

	jsonData := `{"error": "something went wrong"}`

	_, _, err := rp.Parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_Name(t *testing.T) {
	tests := []struct {
		name          string
		processorName string
		responseBody  string
		expected      string
	}{
		{
			name:          "non-empty processor name",
			processorName: "testProcessor",
			responseBody:  "testResponse",
			expected:      "testProcessor",
		},
		{
			name:          "empty processor name, non-empty response body",
			processorName: "",
			responseBody:  "testResponse",
			expected:      "testResponse",
		},
		{
			name:          "empty processor name and response body",
			processorName: "",
			responseBody:  "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{
				name:         tt.processorName,
				responseBody: tt.responseBody,
			}

			result := rp.Name()
			assert.Equal(t, tt.expected, result)
		})
	}
}
