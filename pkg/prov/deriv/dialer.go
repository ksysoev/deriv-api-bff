package deriv

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

type Dialer interface {
	Dial(ctx context.Context, baseUrl string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error)
}

type WebSocketDialer struct {}

func (r *WebSocketDialer) Dial(ctx context.Context, baseUrl string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error) {
	return websocket.Dial(ctx, baseUrl, opts)
}
