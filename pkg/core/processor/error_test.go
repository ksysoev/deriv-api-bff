package processor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		data   json.RawMessage
	}{
		{
			name:   "Valid data",
			data:   []byte(`{"code": "400", "message": "Bad Request", "details": {"field": "value"}}`),
			errMsg: "Bad Request",
		},
		{
			name:   "Invalid data format",
			data:   []byte(`invalid data`),
			errMsg: "failed to unmarshal APIError data: invalid character 'i' looking for beginning of value",
		},
		{
			name:   "Missing code",
			data:   []byte(`{"message": "Bad Request"}`),
			errMsg: "error message is not in expected format",
		},
		{
			name:   "Missing message",
			data:   []byte(`{"code": "400"}`),
			errMsg: "error message is not in expected format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.data)
			assert.EqualError(t, err, tt.errMsg)
		})
	}
}
