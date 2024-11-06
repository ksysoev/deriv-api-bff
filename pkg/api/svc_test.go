package api

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	wasabi "github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSevice(t *testing.T) {
	cfg := &Config{
		Listen: "localhost:8080",
	}

	mockBFFService := NewMockBFFService(t)

	svc := NewSevice(cfg, mockBFFService)

	assert.NotNil(t, svc)
	assert.Equal(t, cfg, svc.cfg)
	assert.Equal(t, mockBFFService, svc.handler)
}

func TestSvc_Run(t *testing.T) {
	config := &Config{
		Listen: ":0",
	}

	service := NewSevice(config, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	defer cancel()

	done := make(chan struct{})

	// Run the server
	go func() {
		err := service.Run(ctx)
		switch err {
		case nil:
			close(done)
		default:
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to stop")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		expected wasabi.Request
		name     string
		data     []byte
		msgType  wasabi.MessageType
	}{
		{
			name:     "TextMessage",
			msgType:  wasabi.MsgTypeText,
			data:     []byte("test text message"),
			expected: request.NewRequest(context.Background(), request.TextMessage, []byte("test text message")),
		},
		{
			name:     "BinaryMessage",
			msgType:  wasabi.MsgTypeBinary,
			data:     []byte{0x01, 0x02, 0x03},
			expected: request.NewRequest(context.Background(), request.BinaryMessage, []byte{0x01, 0x02, 0x03}),
		},
		{
			name:     "UnsupportedMessageType",
			msgType:  wasabi.MessageType(999),
			data:     []byte("unsupported message"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := parse(nil, context.Background(), tt.msgType, tt.data)
			assert.Equal(t, tt.expected, req)
		})
	}
}
func TestService_Handle(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	mockBFFService := NewMockBFFService(t)
	service := NewSevice(&Config{}, mockBFFService)

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

func TestService_Addr(t *testing.T) {
	cfg := &Config{
		Listen: "localhost:0",
	}

	mockBFFService := NewMockBFFService(t)
	service := NewSevice(cfg, mockBFFService)

	addr := service.Addr()
	assert.Nil(t, addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})

	go func() {
		for service.Addr() == nil {
			time.Sleep(10 * time.Millisecond)
		}

		close(ready)
	}()

	go func() {
		_ = service.Run(ctx)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to start")
	}

	addr = service.Addr()
	assert.NotNil(t, addr)
}
func TestPopulateDefaults(t *testing.T) {
	tests := []struct {
		input    *Config
		expected *Config
		name     string
	}{
		{
			name: "Defaults Applied",
			input: &Config{
				Listen: "localhost:8080",
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        maxRequestsDefault,
				MaxRequestsPerConn: maxRequestsPerConnDefault,
			},
		},
		{
			name: "MaxRequests Set",
			input: &Config{
				Listen:      "localhost:8080",
				MaxRequests: 200,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: maxRequestsPerConnDefault,
			},
		},
		{
			name: "MaxRequestsPerConn Set",
			input: &Config{
				Listen:             "localhost:8080",
				MaxRequestsPerConn: 20,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        maxRequestsDefault,
				MaxRequestsPerConn: 20,
			},
		},
		{
			name: "All Values Set",
			input: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: 20,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			populateDefaults(tt.input)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestSkipMetrics(t *testing.T) {
	tests := []struct {
		name       string
		routingKey string
		want       bool
	}{
		{
			name:       "TextMessage routing key",
			routingKey: request.TextMessage,
			want:       true,
		},
		{
			name:       "BinaryMessage routing key",
			routingKey: request.BinaryMessage,
			want:       true,
		},
		{
			name:       "Other routing key",
			routingKey: "other",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := mocks.NewMockRequest(t)
			mockRequest.EXPECT().RoutingKey().Return(tt.routingKey)

			got := skipMetrics(mockRequest)
			assert.Equal(t, tt.want, got)
		})
	}
}
