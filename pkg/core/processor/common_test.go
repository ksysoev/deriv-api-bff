package processor

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterResp(t *testing.T) {
	tests := []struct {
		resp     map[string]json.RawMessage
		fieldMap map[string]string
		expected map[string]json.RawMessage
		name     string
		allow    []string
	}{
		{
			name: "filter with allowed keys",
			resp: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
				"key2": json.RawMessage(`"value2"`),
				"key3": json.RawMessage(`"value3"`),
			},
			allow: []string{"key1", "key3"},
			expected: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
				"key3": json.RawMessage(`"value3"`),
			},
		},
		{
			name: "filter with field map",
			resp: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
				"key2": json.RawMessage(`"value2"`),
			},
			allow: []string{"key1", "key2"},
			fieldMap: map[string]string{
				"key1": "newKey1",
				"key2": "newKey2",
			},
			expected: map[string]json.RawMessage{
				"newKey1": json.RawMessage(`"value1"`),
				"newKey2": json.RawMessage(`"value2"`),
			},
		},
		{
			name: "filter with missing keys",
			resp: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
			},
			allow: []string{"key1", "key2"},
			expected: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
			},
		},
		{
			name: "filter with empty allow list",
			resp: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
				"key2": json.RawMessage(`"value2"`),
			},
			allow:    []string{},
			expected: map[string]json.RawMessage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterResp(tt.resp, tt.allow, tt.fieldMap)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestPrepareResp(t *testing.T) {
	tests := []struct {
		err      error
		expected map[string]json.RawMessage
		name     string
		data     []byte
	}{
		{
			name: "valid JSON object",
			data: []byte(`{"key1":"value1","key2":"value2"}`),
			expected: map[string]json.RawMessage{
				"key1": json.RawMessage(`"value1"`),
				"key2": json.RawMessage(`"value2"`),
			},
			err: nil,
		},
		{
			name: "valid JSON array",
			data: []byte(`["value1","value2"]`),
			expected: map[string]json.RawMessage{
				"list": json.RawMessage(`["value1","value2"]`),
			},
			err: nil,
		},
		{
			name: "valid JSON value",
			data: []byte(`"value1"`),
			expected: map[string]json.RawMessage{
				"value": json.RawMessage(`"value1"`),
			},
			err: nil,
		},
		{
			name:     "empty data",
			data:     []byte(``),
			expected: nil,
			err:      fmt.Errorf("response body not found"),
		},
		{
			name:     "invalid JSON",
			data:     []byte(`{key1:value1}`),
			expected: nil,
			err:      fmt.Errorf("failed to unmarshal response body: invalid character 'k' looking for beginning of object key string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prepareResp(tt.data)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
