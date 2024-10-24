package deriv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

var wsHandlerEcho = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	for {
		_, wsr, err := c.Reader(r.Context())
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}

		wsw, err := c.Writer(r.Context(), websocket.MessageText)
		if err != nil {
			return
		}

		_, err = io.Copy(wsw, wsr)
		if err != nil {
			return
		}

		err = wsw.Close()
		if err != nil {
			return
		}
	}
})

func TestService_dialer(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		expectedURL   string
		expectedError string
	}{
		{
			name:          "Missing app_id",
			baseURL:       "wss://example.com",
			expectedError: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			s := &Service{}
			conn, err := s.dialer(ctx, tt.baseURL)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
			}
		})
	}
}

func TestService_createMessage(t *testing.T) {
	tests := []struct {
		name          string
		requestType   string
		expectedError string
		requestData   []byte
		expectedData  []byte
		expectedType  wasabi.MessageType
	}{
		{
			name:         "TextMessage",
			requestType:  request.TextMessage,
			requestData:  []byte("text data"),
			expectedType: wasabi.MsgTypeText,
			expectedData: []byte("text data"),
		},
		{
			name:         "BinaryMessage",
			requestType:  request.BinaryMessage,
			requestData:  []byte{0x01, 0x02, 0x03},
			expectedType: wasabi.MsgTypeBinary,
			expectedData: []byte{0x01, 0x02, 0x03},
		},
		{
			name:          "UnsupportedMessage",
			requestType:   "unsupported",
			requestData:   nil,
			expectedError: "unsupported request type: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}

			req := mocks.NewMockRequest(t)
			req.EXPECT().RoutingKey().Return(tt.requestType)

			if tt.requestData != nil {
				req.EXPECT().Data().Return(tt.requestData)
			}

			msgType, data, err := s.createMessage(req)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, msgType)
				assert.Equal(t, tt.expectedData, data)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedURL string
	}{
		{
			name: "Valid config",
			config: &Config{
				Endpoint: "wss://example.com",
			},
			expectedURL: "wss://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.config)

			assert.NotNil(t, service)
			assert.NotNil(t, service.handler)

			// Assuming backend.NewWSBackend returns a struct with an Endpoint field for testing purposes
			_, ok := service.handler.(*backend.WSBackend)
			assert.True(t, ok)
		})
	}
}
func TestService_Handle(t *testing.T) {
	tests := []struct {
		name          string
		conn          *core.Conn
		req           *request.Request
		handlerReturn error
		expectedError string
	}{
		{
			name:          "Handler returns no error",
			conn:          &core.Conn{},
			req:           &request.Request{},
			handlerReturn: nil,
		},
		{
			name:          "Handler returns an error",
			conn:          &core.Conn{},
			req:           &request.Request{},
			handlerReturn: fmt.Errorf("handler error"),
			expectedError: "handler error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := mocks.NewMockBackend(t)
			mockHandler.EXPECT().Handle(tt.conn, tt.req).Return(tt.handlerReturn)

			s := &Service{
				handler: mockHandler,
			}

			err := s.Handle(tt.conn, tt.req)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockHandler.AssertExpectations(t)
		})
	}
}
func TestService_dial(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()

	baseURL := "ws://" + server.Listener.Addr().String()

	tests := []struct {
		name          string
		baseURL       string
		urlParams     url.Values
		headers       http.Header
		expectedURL   string
		expectedError string
	}{
		{
			name:          "Missing app_id",
			baseURL:       baseURL,
			urlParams:     url.Values{},
			expectedError: "app_id is required",
		},
		{
			name:    "Valid app_id and lang",
			baseURL: baseURL,
			urlParams: url.Values{
				"app_id": []string{"123"},
				"l":      []string{"en"},
			},
			headers:     http.Header{"Authorization": []string{"Bearer token"}},
			expectedURL: "wss://example.com?app_id=123&l=en",
		},
		{
			name:    "Valid app_id without lang",
			baseURL: baseURL,
			urlParams: url.Values{
				"app_id": []string{"123"},
			},
			headers:     http.Header{"Authorization": []string{"Bearer token"}},
			expectedURL: "wss://example.com?app_id=123",
		},
		{
			name:    "invalid URL",
			baseURL: "invalid",
			urlParams: url.Values{
				"app_id": []string{"123"},
			},
			expectedError: "failed to WebSocket dial: unexpected url scheme: \"\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			s := &Service{}
			conn, err := s.dial(ctx, tt.baseURL, tt.urlParams, tt.headers)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
			}
		})
	}
}
