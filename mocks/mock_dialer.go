package mocks

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

type MockWebSocketDialer struct{}

func (r *MockWebSocketDialer) Dial(ctx context.Context, baseUrl string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error) {
	return &websocket.Conn{}, &http.Response{
		Body: nil,
	}, nil
}