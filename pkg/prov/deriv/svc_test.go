package deriv

import (
	"context"
	"errors"
	"net/url"
	"testing"

	mocks "github.com/ksysoev/deriv-api-bff/mocks"
	core "github.com/ksysoev/deriv-api-bff/pkg/core"
	middleware "github.com/ksysoev/deriv-api-bff/pkg/middleware"
	wasabi_mocks "github.com/ksysoev/wasabi/mocks"
)

func TestSvc_Handle_Success(t *testing.T) {
	mockHandler := wasabi_mocks.NewMockRequestHandler(t)
	server := &Service{
		handler: mockHandler,
	}

	mockConn := wasabi_mocks.NewMockConnection(t)
	mockReq := wasabi_mocks.NewMockRequest(t)

	mockHandler.EXPECT().Handle(mockConn, mockReq).Return(nil)

	err := server.Handle(mockConn, mockReq)

	if err != nil {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestSvc_Handle_Error(t *testing.T) {
	mockHandler := wasabi_mocks.NewMockRequestHandler(t)
	server := &Service{
		handler: mockHandler,
	}

	mockConn := wasabi_mocks.NewMockConnection(t)
	mockReq := wasabi_mocks.NewMockRequest(t)
	testErr := errors.New("test")

	mockHandler.EXPECT().Handle(mockConn, mockReq).Return(testErr)

	err := server.Handle(mockConn, mockReq)

	if err != testErr {
		t.Errorf("got unexpected error: %s, expected: %s", err, testErr)
	}
}

func TestSvc_NewService_NilUrlParams(t *testing.T) {
	config := &Config{
		Endpoint: "/",
	}

	service := NewService(config, &WebSocketDialer{})
	ctx := context.Background()
	testErr := errors.New("url params are required")

	mockConn := wasabi_mocks.NewMockConnection(t)
	mockReq := wasabi_mocks.NewMockRequest(t)

	mockConn.EXPECT().ID().Return("conn_id")
	mockConn.EXPECT().Context().Return(ctx)

	err := service.Handle(mockConn, mockReq)

	if err.Error() != testErr.Error() {
		t.Errorf("got unexpected error: %s, but expected: %s", err, testErr)
	}
}

func TestSvc_NewService(t *testing.T) {
	config := &Config{
		Endpoint: "/",
	}

	service := NewService(config, &mocks.MockWebSocketDialer{})
	queryParams := url.Values{}
	queryParams.Set("app_id", "app123")

	var keyQuery middleware.ContextKey = 1
	ctx := context.WithValue(context.Background(), keyQuery, queryParams)

	mockConn := wasabi_mocks.NewMockConnection(t)
	mockReq := wasabi_mocks.NewMockRequest(t)

	mockConn.EXPECT().ID().Return("conn_id")
	mockConn.EXPECT().Context().Return(ctx)
	mockReq.EXPECT().RoutingKey().Return(core.TextMessage)
	mockReq.EXPECT().Context().Return(ctx)
	mockReq.EXPECT().Data().Return([]byte("test request"))

	err := service.Handle(mockConn, mockReq)

	if err != nil {
		t.Errorf("got unexpected error: %s", err)
	}
}