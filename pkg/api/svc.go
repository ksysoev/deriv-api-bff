package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	ProcessReuest(clientConn wasabi.Connection, req *core.Request) error
}

type Config struct {
	Listen string `mapstructure:"listen"`
}

type Service struct {
	cfg     *Config
	handler BFFService
}

func NewSevice(cfg *Config, handler BFFService) *Service {
	return &Service{
		cfg:     cfg,
		handler: handler,
	}
}

func (s *Service) Run(ctx context.Context) error {
	dispatcher := dispatch.NewRouterDispatcher(s, parse)
	registry := channel.NewConnectionRegistry(
		channel.WithMaxFrameLimit(maxMessageSize),
	)
	endpoint := channel.NewChannel("/", dispatcher, registry, channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())
	server := server.NewServer(s.cfg.Listen)
	server.AddChannel(endpoint)

	go func() {
		<-ctx.Done()

		if err := server.Close(); err != nil {
			slog.Error("Fail to close app server", "error", err)
		}
	}()

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}
	return nil
}

func (s *Service) Handle(conn wasabi.Connection, r wasabi.Request) error {
	req, ok := r.(*core.Request)
	if !ok {
		return fmt.Errorf("unsupported request type: %T", req)
	}

	switch req.RoutingKey() {
	case core.TextMessage, core.BinaryMessage:
		return s.handler.PassThrough(conn, req)
	case "":
		return fmt.Errorf("Empty request type: %v", req)
	default:
		return s.handler.ProcessReuest(conn, req)
	}
}

func parse(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
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
