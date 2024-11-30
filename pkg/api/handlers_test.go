package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	request "github.com/ksysoev/deriv-api-bff/pkg/core/request"
	wasabi "github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestService_Handle(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	mockBFFService := NewMockBFFService(t)
	service, err := NewSevice(&Config{}, mockBFFService)

	assert.NoError(t, err)

	mockBFFService.EXPECT().PassThrough(mockConn, mock.Anything).Return(nil)
	mockBFFService.EXPECT().ProcessRequest(mockConn, mock.Anything).Return(nil)

	tests := []struct {
		request     wasabi.Request
		name        string
		expectError bool
	}{
		{
			name:        "PassThrough TextMessage",
			request:     request.NewRequest(context.Background(), request.TextMessage, []byte("test text message")),
			expectError: false,
		},
		{
			name:        "PassThrough BinaryMessage",
			request:     request.NewRequest(context.Background(), request.BinaryMessage, []byte{0x01, 0x02, 0x03}),
			expectError: false,
		},
		{
			name:        "Empty Request Type",
			request:     request.NewRequest(context.Background(), "", []byte("empty request type")),
			expectError: true,
		},
		{
			name:        "ProcessRequest CustomMessage",
			request:     request.NewRequest(context.Background(), "customMessage", []byte("custom message")),
			expectError: false,
		},
		{
			name:        "Unsupported Request Type",
			request:     mocks.NewMockRequest(t),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Handle(mockConn, tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_HealthCheck(t *testing.T) {
	service := &Service{}

	req, err := http.NewRequest("GET", "/livez", http.NoBody)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(service.HealthCheck)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}
