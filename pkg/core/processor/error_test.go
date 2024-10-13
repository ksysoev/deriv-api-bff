package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name   string
		data   any
		errMsg string
	}{
		{
			name:   "Valid data",
			data:   map[string]any{"code": "400", "message": "Bad Request", "details": map[string]any{"field": "value"}},
			errMsg: "Bad Request",
		},
		{
			name:   "Invalid data format",
			data:   "invalid data",
			errMsg: "error data is not in expected format",
		},
		{
			name:   "Missing code",
			data:   map[string]any{"message": "Bad Request"},
			errMsg: "error code is not in expected format",
		},
		{
			name:   "Missing message",
			data:   map[string]any{"code": "400"},
			errMsg: "error message is not in expected format",
		},
		{
			name:   "Invalid details",
			data:   map[string]any{"code": "400", "message": "Bad Request", "details": make(chan int)},
			errMsg: "failed to marshal APIError details: json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.data)
			assert.EqualError(t, err, tt.errMsg)
		})
	}
}
