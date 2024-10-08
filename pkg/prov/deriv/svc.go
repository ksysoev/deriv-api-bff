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
	dialer Dialer
}

// NewService initializes and returns a new Service instance.
// It takes cfg of type *Config which contains configuration settings.
// It returns a pointer to a Service struct.
func NewService(cfg *Config, dialer Dialer) *Service {
	s := &Service{
		dialer: dialer,
	}

	s.handler = backend.NewWSBackend(
		cfg.Endpoint,
		s.createMessage,
		backend.WithWSDialler(s.dialer),
	)

	return s
}

// Handle processes a request using the provided connection and request objects.
// It takes conn of type *core.Conn and req of type *core.Request.
// It returns an error if the handler fails to process the request.
func (s *Service) Handle(conn *core.Conn, req *core.Request) error {
	return s.handler.Handle(conn, req)
}

// createMessage constructs a wasabi.MessageType and its corresponding byte data from a wasabi.Request.
// It takes a single parameter r of type wasabi.Request.
// It returns a wasabi.MessageType, a byte slice containing the message data, and an error if the request type is unsupported.
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

// dialer establishes a WebSocket connection to the specified baseURL using context-derived parameters.
// It takes ctx of type context.Context and baseURL of type string.
// It returns a pointer to websocket.Conn and an error.
// It returns an error if the connection cannot be established or if there are issues with the provided context parameters.
func (s *Service) dialer(ctx context.Context, baseURL string) (*websocket.Conn, error) {
	urlParams := middleware.QueryParamsFromContext(ctx)
	headers := middleware.HeadersFromContext(ctx)

	return s.dial(ctx, baseURL, urlParams, headers)
}

// dial establishes a WebSocket connection to the specified baseURL with the given URL parameters and headers.
// It takes ctx of type context.Context, baseURL of type string, urlParams of type url.Values, and headers of type http.Header.
// It returns a pointer to a websocket.Conn and an error.
// It returns an error if the app_id parameter is missing or if the WebSocket connection fails.
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

	c, resp, err := s.dialer.Dial(ctx, baseURL, &websocket.DialOptions{
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
