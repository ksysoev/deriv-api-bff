package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/server"
)

const (
	maxMessageSize = 600 * 1024
)

type BFFService interface {
	PassThrough(clientConn wasabi.Connection, req *core.Request) error
	ProcessRequest(clientConn wasabi.Connection, req *core.Request) error
}

type Config struct {
	Listen string `mapstructure:"listen"`
}

type Service struct {
	cfg     *Config
	handler BFFService
}

// NewSevice creates a new instance of Service with the provided configuration and handler.
// It takes cfg of type *Config and handler of type BFFService.
// It returns a pointer to a Service struct.
func NewSevice(cfg *Config, handler BFFService) *Service {
	return &Service{
		cfg:     cfg,
		handler: handler,
	}
}

// Run starts the service and listens for incoming connections.
// It takes a context.Context parameter which is used to manage the lifecycle of the service.
// It returns an error if the server fails to start or close properly.
// The function sets up a dispatcher, a connection registry, and a channel endpoint with middleware.
// It also handles graceful shutdown when the context is done.
func (s *Service) Run(ctx context.Context) error {
	dispatcher := dispatch.NewRouterDispatcher(s, parse)
	registry := channel.NewConnectionRegistry(
		channel.WithMaxFrameLimit(maxMessageSize),
	)
	endpoint := channel.NewChannel("/", dispatcher, registry, channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())

	srv := server.NewServer(s.cfg.Listen)
	srv.AddChannel(endpoint)

	go func() {
		<-ctx.Done()

		if err := srv.Close(); err != nil {
			slog.Error("Fail to close app server", "error", err)
		}
	}()

	if err := srv.Run(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Handle processes a request received on a connection and routes it based on the request type.
// It takes conn of type wasabi.Connection and r of type wasabi.Request.
// It returns an error if the request type is unsupported or if the request type is empty.
// If the request type is core.TextMessage or core.BinaryMessage, it passes the request through to the handler.
// For other request types, it processes the request using the handler.
func (s *Service) Handle(conn wasabi.Connection, r wasabi.Request) error {
	req, ok := r.(*core.Request)
	if !ok {
		return fmt.Errorf("unsupported request type: %T", req)
	}

	switch req.RoutingKey() {
	case core.TextMessage, core.BinaryMessage:
		return s.handler.PassThrough(conn, req)
	case "":
		return fmt.Errorf("empty request type: %v", req)
	default:
		return s.handler.ProcessRequest(conn, req)
	}
}

// parse processes a message received over a Wasabi connection and converts it into a core request.
// It takes conn of type wasabi.Connection, ctx of type context.Context, msgType of type wasabi.MessageType, and data of type []byte.
// It returns a wasabi.Request which represents the parsed message.
// If the msgType is unsupported, it logs an error and returns nil.
func parse(_ wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request { //nolint:revive //Defined by Wasabi
	var coreMsgType string

	switch msgType {
	case wasabi.MsgTypeText:
		coreMsgType = core.TextMessage
	case wasabi.MsgTypeBinary:
		coreMsgType = core.BinaryMessage
	default:
		slog.Error("unsupported message type", "type", msgType)
		return nil
	}

	return core.NewRequest(ctx, coreMsgType, data)
}
