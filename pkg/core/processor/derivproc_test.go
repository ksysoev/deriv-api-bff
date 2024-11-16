package processor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/tmpl"
	"github.com/stretchr/testify/assert"
)

func TestNewDeriv(t *testing.T) {
	tests := []struct {
		cfg     *Config
		name    string
		wantErr bool
	}{
		{
			name: "Valid Deriv Config",
			cfg: &Config{
				Request:  map[string]any{"params": "${params}", "req_id": "${req_id}"},
				FieldMap: map[string]string{"key1": "mappedKey1"},
				Allow:    []string{"key1", "key2"},
			},
			wantErr: false,
		},
		{
			name: "Fail to marshal request template",
			cfg: &Config{
				Request:  map[string]any{"params": "${params}", "data": make(chan int)},
				FieldMap: map[string]string{"key1": "mappedKey1"},
				Allow:    []string{"key1", "key2"},
			},
			wantErr: true,
		},
		{
			name: "Fail to parse request template",
			cfg: &Config{
				Request:  map[string]any{"test": "${params}invalid"},
				FieldMap: map[string]string{"key1": "mappedKey1"},
				Allow:    []string{"key1", "key2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := NewDeriv(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, processor)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, processor)
			assert.NotNil(t, processor.tmpl)
			assert.Equal(t, tt.cfg.FieldMap, processor.fieldMap)
			assert.Equal(t, tt.cfg.Allow, processor.allow)
		})
	}
}

func TestProcessor_Render(t *testing.T) {
	tmpl1 := tmpl.MustNewTmpl(`{"params":"${params}","req_id":"${req_id}","resp": "${resp}"}`)
	tmpl2 := tmpl.MustNewTmpl(`{"params": "${params.key1.key2}}"}`)

	tests := []struct {
		deps     map[string]any
		tmpl     *tmpl.Tmpl
		name     string
		expected string
		reqID    string
		params   []byte
		wantErr  bool
	}{
		{
			name:     "with params and deps",
			params:   []byte(`{"key1": "value1"}`),
			deps:     map[string]any{"dep1": "value1"},
			reqID:    "12345",
			expected: `{"params":{"key1":"value1"},"req_id":"12345","resp": {"dep1":"value1"}}`,
			wantErr:  false,
		},
		{
			name:     "with nil params and deps",
			params:   nil,
			deps:     nil,
			reqID:    "12345",
			expected: `{"params":{},"req_id":"12345","resp": {}}`,
			wantErr:  false,
		},
		{
			name:     "with empty params and deps",
			params:   []byte(`{}`),
			deps:     map[string]any{},
			reqID:    "12345",
			expected: `{"params":{},"req_id":"12345","resp": {}}`,
			wantErr:  false,
		},
		{
			name:     "with only params",
			params:   []byte(`{"key1": "value1"}`),
			deps:     nil,
			reqID:    "12345",
			expected: `{"params":{"key1":"value1"},"req_id":"12345","resp": {}}`,
			wantErr:  false,
		},
		{
			name:     "with only deps",
			params:   nil,
			deps:     map[string]any{"dep1": "value1"},
			reqID:    "12345",
			expected: `{"params":{},"req_id":"12345","resp": {"dep1":"value1"}}`,
			wantErr:  false,
		},
		{
			name:     "with error",
			params:   nil,
			deps:     nil,
			reqID:    "12345",
			tmpl:     tmpl2,
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{
				tmpl: tmpl1,
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
		expected map[string]json.RawMessage
		name     string
		jsonData string
	}{
		{
			name:     "object",
			jsonData: `{"data": {"key1": "value1", "key2": "value2"}, "msg_type": "data"}`,
			expected: map[string]json.RawMessage{"key1": []byte(`"value1"`), "key2": []byte(`"value2"`)},
		},
		{
			name:     "array",
			jsonData: `{"data": [{"key1": "value1"}, {"key2": "value2"}], "msg_type": "data"}`,
			expected: map[string]json.RawMessage{"list": []byte(`[{"key1": "value1"}, {"key2": "value2"}]`)},
		},
		{
			name:     "scalar",
			jsonData: `{"data": "value", "msg_type": "data"}`,
			expected: map[string]json.RawMessage{"value": []byte(`"value"`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{}

			result, err := rp.parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_parse_Error(t *testing.T) {
	rp := &DerivProc{}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_ResponseBodyNotFound(t *testing.T) {
	rp := &DerivProc{}

	jsonData := `{"key": "value"}`

	_, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_parse_UnexpectedFormat(t *testing.T) {
	rp := &DerivProc{}

	jsonData := `{invalid json}`

	result, err := rp.parse([]byte(jsonData))
	assert.Error(t, err)
	assert.Nil(t, result)
}
func TestProcessor_Parse_Success(t *testing.T) {
	tests := []struct {
		fieldMap map[string]string
		expected map[string]any
		name     string
		jsonData string
		allow    []string
	}{
		{
			name:     "allowed fields with field mapping",
			fieldMap: map[string]string{"key1": "mappedKey1"},
			allow:    []string{"key1", "key2"},
			jsonData: `{"data": {"key1": "value1", "key2": "value2"}, "msg_type": "data"}`,
			expected: map[string]any{"mappedKey1": "value1", "key2": "value2"},
		},
		{
			name:     "allowed fields without field mapping",
			fieldMap: nil,
			allow:    []string{"key1", "key2"},
			jsonData: `{"data": {"key1": "value1", "key2": "value2"}, "msg_type": "data"}`,
			expected: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "missing allowed fields",
			fieldMap: nil,
			allow:    []string{"key3"},
			jsonData: `{"data": {"key1": "value1", "key2": "value2"}, "msg_type": "data"}`,
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{
				fieldMap: tt.fieldMap,
				allow:    tt.allow,
			}

			resp, err := rp.Parse([]byte(tt.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, resp.Filtered())
		})
	}
}

func TestProcessor_Parse_Error(t *testing.T) {
	rp := &DerivProc{}

	jsonData := `{"error": "something went wrong"}`

	_, err := rp.Parse([]byte(jsonData))
	assert.Error(t, err)
}

func TestProcessor_Name(t *testing.T) {
	tests := []struct {
		name          string
		processorName string
		expected      string
	}{
		{
			name:          "non-empty processor name",
			processorName: "testProcessor",
			expected:      "testProcessor",
		},
		{
			name:          "empty processor name",
			processorName: "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &DerivProc{
				name: tt.processorName,
			}

			result := rp.Name()
			assert.Equal(t, tt.expected, result)
		})
	}
}
