package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	err := NewAPIError("code", "message", json.RawMessage([]byte(`{"key": "value"}`)))

	assert.Equal(t, "code", err.Code)
	assert.Equal(t, "message", err.Message)
	assert.Equal(t, json.RawMessage([]byte(`{"key": "value"}`)), err.Details)
}

func TestAPIError_Error(t *testing.T) {
	err := NewAPIError("code", "message", json.RawMessage([]byte(`{"key": "value"}`)))

	assert.Equal(t, "message", err.Error())
}

func TestAPIError_Encode(t *testing.T) {
	err := NewAPIError("code", "message", json.RawMessage([]byte(`{"key": "value"}`)))

	assert.Equal(t, json.RawMessage([]byte(`{"code":"code","message":"message","details":{"key":"value"}}`)), err.Encode())
}
