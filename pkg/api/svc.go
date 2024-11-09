package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	reqmid "github.com/ksysoev/wasabi/middleware/request"
	"github.com/ksysoev/wasabi/server"
)

const (
	maxMessageSize                  = 600 * 1024
	maxRequestsDefault              = 100
	maxRequestsPerConnDefault       = 10
	generalRateLimitIntervalDefault = "1m"
	generalRateLimitDefault         = 100
)

type BFFService interface {
	PassThrough(clientConn wasabi.Connection, req *request.Request) error
	ProcessRequest(clientConn wasabi.Connection, req *request.Request) error
}

type Config struct {
	Listen             string     `mapstructure:"listen"`
	RateLimits         RateLimits `mapstructure:"rate_limits"`
	MaxRequests        uint       `mapstructure:"max_requests"`
	MaxRequestsPerConn uint       `mapstructure:"max_requests_per_conn"`
}

type RateLimits struct {
	Groups  GroupRateLimits   `mapstructure:"groups"`
	General GeneralRateLimits `mapstructure:"general"`
}

type GeneralRateLimits struct {
	Interval string `mapstructure:"interval"`
	Limit    uint   `mapstructure:"limit"`
}

type GroupRateLimits struct {
}

type Service struct {
	cfg     *Config
	handler BFFService
	server  *server.Server
}

// NewSevice creates a new instance of Service with the provided configuration and handler.
// It takes cfg of type *Config and handler of type BFFService.
// It returns a pointer to a Service struct.
func NewSevice(cfg *Config, handler BFFService) *Service {
	s := &Service{
		cfg:     cfg,
		handler: handler,
	}

	populateDefaults(cfg)

	dispatcher := dispatch.NewRouterDispatcher(s, parse)
	dispatcher.Use(middleware.NewErrorHandlingMiddleware())
	dispatcher.Use(middleware.NewMetricsMiddleware("bff-deriv", skipMetrics))
	dispatcher.Use(reqmid.NewTrottlerMiddleware(cfg.MaxRequests))
	dispatcher.Use(reqmid.NewRateLimiterMiddleware(getRequestLimits(cfg.RateLimits)))

	registry := channel.NewConnectionRegistry(
		channel.WithMaxFrameLimit(maxMessageSize),
		channel.WithConcurrencyLimit(cfg.MaxRequestsPerConn),
	)
	endpoint := channel.NewChannel("/", dispatcher, registry, channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())

	s.server = server.NewServer(cfg.Listen)
	s.server.AddChannel(endpoint)

	return s
}

// Addr returns the network address the server is listening on.
// It takes no parameters.
// It returns a net.Addr which represents the server's network address.
func (s *Service) Addr() net.Addr {
	return s.server.Addr()
}

// Run starts the service and listens for incoming connections.
// It takes a context.Context parameter which is used to manage the lifecycle of the service.
// It returns an error if the server fails to start or close properly.
// The function sets up a dispatcher, a connection registry, and a channel endpoint with middleware.
// It also handles graceful shutdown when the context is done.
func (s *Service) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		if err := s.server.Close(); err != nil {
			slog.Error("Fail to close app server", "error", err)
		}
	}()

	if err := s.server.Run(); err != nil {
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
	req, ok := r.(*request.Request)
	if !ok {
		return fmt.Errorf("unsupported request type: %T", req)
	}

	switch req.RoutingKey() {
	case request.TextMessage, request.BinaryMessage:
		return s.handler.PassThrough(conn, req)
	case "":
		return fmt.Errorf("empty request type: %v", req)
	default:
		return s.handler.ProcessRequest(conn, req)
	}
}

// getRequestLimits is a helper function transform request limits mentioned in server configuration
// into wasabi request limits, so that we can use the RateLimiter middleware.
func getRequestLimits(rateLimitCfg RateLimits) func(wasabi.Request) (string, time.Duration, uint64) {
	return func(r wasabi.Request) (string, time.Duration, uint64) {
		duration, err := time.ParseDuration(rateLimitCfg.General.Interval)
		if err != nil {
			//TODO: should we panic or fallback to defaults with warning?
			panic(errors.New("incorrect interval format"))
		}

		return r.RoutingKey(), duration, uint64(rateLimitCfg.General.Limit)
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
		coreMsgType = request.TextMessage
	case wasabi.MsgTypeBinary:
		coreMsgType = request.BinaryMessage
	default:
		slog.Error("unsupported message type", "type", msgType)
		return nil
	}

	return request.NewRequest(ctx, coreMsgType, data)
}

// populateDefaults sets default values for the configuration if they are not already set.
// It takes a single parameter cfg of type *Config.
// It does not return any values.
func populateDefaults(cfg *Config) {
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = maxRequestsDefault
	}

	if cfg.MaxRequestsPerConn == 0 {
		cfg.MaxRequestsPerConn = maxRequestsPerConnDefault
	}

	if cfg.RateLimits.General.Interval == "" {
		cfg.RateLimits.General.Interval = generalRateLimitIntervalDefault
	}

	if cfg.RateLimits.General.Limit == 0 {
		cfg.RateLimits.General.Limit = generalRateLimitDefault
	}
}

// skipMetrics determines whether metrics should be skipped for a given request.
// It takes a single parameter r of type wasabi.Request.
// It returns a boolean value: true if the request's routing key is either TextMessage or BinaryMessage, otherwise false.
func skipMetrics(r wasabi.Request) bool {
	if r.RoutingKey() == request.TextMessage || r.RoutingKey() == request.BinaryMessage {
		return true
	}

	return false
}
