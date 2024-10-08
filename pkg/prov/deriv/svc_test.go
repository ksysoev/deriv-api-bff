package deriv

import (
	"context"
	"errors"
	"testing"

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
