package deriv

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
)

const (
	maxMessageSize = 600 * 1024
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
}

type Service struct {
	handler wasabi.RequestHandler
}

func NewService(cfg *Config) *Service {
	s := &Service{}

	s.handler = backend.NewWSBackend(
		cfg.Endpoint,
		s.createMessage,
		backend.WithWSDialler(s.dialer),
	)

	return s
}

func (s *Service) Handle(conn *core.Conn, req *core.Request) error {
	return s.handler.Handle(conn, req)
}

func (s *Service) createMessage(r wasabi.Request) (wasabi.MessageType, []byte, error) {
	switch r.RoutingKey() {
	case core.TextMessage:
		return wasabi.MsgTypeText, r.Data(), nil
	case core.BinaryMessage:
		return wasabi.MsgTypeBinary, r.Data(), nil
	default:
		var t wasabi.MessageType
		return t, nil, fmt.Errorf("unsupported request type: %s", r.RoutingKey())
	}
}

func (s *Service) dialer(ctx context.Context, baseURL string) (*websocket.Conn, error) {
	urlParams := middleware.QueryParamsFromContext(ctx)
	headers := middleware.HeadersFromContext(ctx)

	return s.dial(ctx, baseURL, urlParams, headers)
}

func (s *Service) dial(ctx context.Context, baseURL string, urlParams url.Values, headers http.Header) (*websocket.Conn, error) {
	if urlParams != nil {
		if app_id := urlParams.Get("app_id"); app_id != "" {
			baseURL = fmt.Sprintf("%s?app_id=%s", baseURL, app_id)
		} else {
			return nil, fmt.Errorf("app_id is required")
		}

		if lang := urlParams.Get("l"); lang != "" {
			baseURL = fmt.Sprintf("%s&l=%s", baseURL, lang)
		}
	} else {
		return nil, fmt.Errorf("app_id is required")
	}

	c, resp, err := websocket.Dial(ctx, baseURL, &websocket.DialOptions{
		HTTPHeader: headers,
	})

	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	c.SetReadLimit(maxMessageSize)

	return c, nil
}
