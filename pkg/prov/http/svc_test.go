package http

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	service := NewService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.handler)
}

func TestRequestFactory(t *testing.T) {
	service := &Service{}
	ctx := context.Background()

	tests := []struct {
		name        string
		request     wasabi.Request
		expectError bool
	}{
		{
			name:        "Valid HTTPReq",
			request:     request.NewHTTPReq(ctx, "GET", "http://localhost/", nil),
			expectError: false,
		},
		{
			name:        "Invalid Request Type",
			request:     mocks.NewMockRequest(t),
			expectError: true,
		},
		{
			name:        "HTTPReq with Error",
			request:     request.NewHTTPReq(ctx, "/invalid", "test", nil), // Assuming HTTPReq can be invalid
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.requestFactory(tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
