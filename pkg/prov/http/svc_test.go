package http

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		request     wasabi.Request
		name        string
		expectError bool
	}{
		{
			name:        "Valid HTTPReq",
			request:     request.NewHTTPReq(ctx, "GET", "http://localhost/", nil, 1),
			expectError: false,
		},
		{
			name:        "Invalid Request Type",
			request:     mocks.NewMockRequest(t),
			expectError: true,
		},
		{
			name:        "HTTPReq with Error",
			request:     request.NewHTTPReq(ctx, "/invalid", "test", nil, 1), // Assuming HTTPReq can be invalid
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

func TestService_Handle(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	mockConn := mocks.NewMockConnection(t)
	conn := core.NewConnection(mockConn, func(_ string) {})
	mockRequest := request.NewHTTPReq(ctx, "GET", "http://localhost/", nil, 1)

	mockHandler := mocks.NewMockRequestHandler(t)
	mockHandler.EXPECT().Handle(mock.Anything, mockRequest).Return(assert.AnError)

	service.handler = mockHandler

	err := service.Handle(conn, mockRequest)

	assert.ErrorIs(t, err, assert.AnError)
}
