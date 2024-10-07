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

// NewService initializes and returns a new instance of Service configured with the provided Config.
// It sets up a WebSocket backend handler using the specified endpoint and additional options.
//
// Parameters:
//   - cfg: A pointer to a Config struct containing the necessary configuration settings.
//
// Returns:
//   - A pointer to a newly created Service instance.
func NewService(cfg *Config) *Service {
	s := &Service{}

	s.handler = backend.NewWSBackend(
		cfg.Endpoint,
		s.createMessage,
		backend.WithWSDialler(s.dialer),
	)

	return s
}

// Handle processes an incoming request using the provided connection and request objects.
// It delegates the handling of the request to the underlying handler associated with the Service.
//
// Parameters:
//   - conn: A pointer to a core.Conn object representing the connection through which the request is received.
//   - req: A pointer to a core.Request object containing the details of the incoming request.
//
// Returns:
//   - error: An error object if the request handling fails, otherwise nil.
func (s *Service) Handle(conn *core.Conn, req *core.Request) error {
	return s.handler.Handle(conn, req)
}

// createMessage constructs a Wasabi message based on the provided request.
// It returns the appropriate Wasabi message type and the message data as a byte slice.
// If the request type is unsupported, it returns an error.
//
// Parameters:
//
//	r (wasabi.Request): The request containing the routing key and data.
//
// Returns:
//
//	wasabi.MessageType: The type of the Wasabi message (e.g., text or binary).
//	[]byte: The message data.
//	error: An error if the request type is unsupported, otherwise nil.
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

// dialer establishes a WebSocket connection to the specified baseURL using
// context-derived query parameters and headers. It returns the established
// WebSocket connection or an error if the connection could not be established.
//
// Parameters:
//
//	ctx - The context containing request-scoped values such as query parameters
//	      and headers.
//	baseURL - The base URL to which the WebSocket connection should be made.
//
// Returns:
//
//	*websocket.Conn - The established WebSocket connection.
//	error - An error if the connection could not be established.
func (s *Service) dialer(ctx context.Context, baseURL string) (*websocket.Conn, error) {
	urlParams := middleware.QueryParamsFromContext(ctx)
	headers := middleware.HeadersFromContext(ctx)

	return s.dial(ctx, baseURL, urlParams, headers)
}

// dial establishes a WebSocket connection to the specified baseURL with the provided URL parameters and HTTP headers.
//
// Parameters:
//   - ctx: The context for controlling the lifetime of the WebSocket connection.
//   - baseURL: The base URL to which the WebSocket connection will be established.
//   - urlParams: URL parameters to be appended to the baseURL. The "app_id" parameter is mandatory.
//   - headers: HTTP headers to be included in the WebSocket handshake request.
//
// Returns:
//   - *websocket.Conn: The established WebSocket connection.
//   - error: An error if the connection could not be established or if required parameters are missing.
//
// The function requires the "app_id" parameter to be present in urlParams. If "app_id" is missing, an error is returned.
// Optionally, the "l" parameter can be included in urlParams to specify the language.
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
