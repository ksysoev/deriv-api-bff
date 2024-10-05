package deriv

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
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
		backend.WithWSDialler(s.dial),
	)

	return s
}

func (s *Service) Handle(conn wasabi.Connection, req wasabi.Request) error {
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

func (s *Service) dial(ctx context.Context, baseURL string) (*websocket.Conn, error) {
	urlParams := middleware.QueryParamsFromContext(ctx)

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
		return nil, fmt.Errorf("url params are required")
	}
	header := http.Header{}
	if h := middleware.HeadersFromContext(ctx); h != nil {
		header = h
	}

	c, resp, err := websocket.Dial(ctx, baseURL, &websocket.DialOptions{
		HTTPHeader: header,
	})

	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	return c, nil
}
