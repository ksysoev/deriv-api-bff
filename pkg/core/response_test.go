package core

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/stretchr/testify/assert"
)

func TestCreateResponse(t *testing.T) {
	rawReq := []byte(`{"req_id":1,"method":"testMethod","params":{"key":"value"}}`)
	ctx := context.Background()

	req := request.NewRequest(ctx, request.TextMessage, rawReq)

	resp := map[string]any{"key": "value"}
	data, err := createResponse(req, resp, nil)
	assert.Nil(t, err)

	expected := []byte(`{"echo":{"req_id":1,"method":"testMethod","params":{"key":"value"}},"key":"value","msg_type":"testMethod","req_id":1}`)
	assert.Equal(t, expected, data)
}

func TestCreateResponseWithAPIError(t *testing.T) {
	rawReq := []byte(`{"req_id":1,"method":"testMethod","params":{"key":"value"}}`)
	ctx := context.Background()

	req := request.NewRequest(ctx, request.TextMessage, rawReq)

	apiErr := NewAPIError("BadRequest", "Bad Request", nil)
	data, err := createResponse(req, nil, apiErr)
	assert.Nil(t, err)

	expected := []byte(`{"echo":{"req_id":1,"method":"testMethod","params":{"key":"value"}},"error":{"code":"BadRequest","message":"Bad Request"},"msg_type":"testMethod","req_id":1}`)
	assert.Equal(t, expected, data)
}

func TestCreateResponseError(t *testing.T) {
	rawReq := []byte(`{"req_id":1,"method":"testMethod","params":{"key":"value"}}`)
	ctx := context.Background()

	req := request.NewRequest(ctx, request.TextMessage, rawReq)

	data, err := createResponse(req, nil, assert.AnError)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, data)
}

func TestCreateResponseWithPassthrough(t *testing.T) {
	rawReq := []byte(`{"req_id":1,"method":"testMethod","params":{"key":"value"},"passthrough":"test"}`)
	ctx := context.Background()

	req := request.NewRequest(ctx, request.TextMessage, rawReq)

	resp := map[string]any{"key": "value"}
	data, err := createResponse(req, resp, nil)
	assert.Nil(t, err)

	expected := []byte(`{"echo":{"req_id":1,"method":"testMethod","params":{"key":"value"},"passthrough":"test"},"key":"value","msg_type":"testMethod","passthrough":"test","req_id":1}`)
	assert.Equal(t, expected, data)
}
