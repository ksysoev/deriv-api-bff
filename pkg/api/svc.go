package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	httpmid "github.com/ksysoev/wasabi/middleware/http"
	reqmid "github.com/ksysoev/wasabi/middleware/request"
	"github.com/ksysoev/wasabi/server"
)

const (
	maxMessageSize                  = 600 * 1024
	maxRequestsDefault              = 100
	maxRequestsPerConnDefault       = 10
	generalRateLimitIntervalDefault = "1m"
	generalRateLimitDuration        = 1 * time.Millisecond
	generalRateLimitDefault         = 100000
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
	Groups  []GroupRateLimits `mapstructure:"groups"`
	General GeneralRateLimits `mapstructure:"general"`
}

type GeneralRateLimits struct {
	Interval string `mapstructure:"interval"`
	Limit    uint   `mapstructure:"limit"`
}

type GroupRateLimits struct {
	Name    string            `mapstructure:"name"`
	Limits  GeneralRateLimits `mapstructure:"limits"`
	Methods []string          `mapstructure:"methods"`
}

type Service struct {
	cfg     *Config
	handler BFFService
	server  *server.Server
}

type groupRatesMapType map[string]struct {
	Name     string
	Methods  []string
	Interval time.Duration
	Limit    uint
}

// NewSevice creates a new instance of Service with the provided configuration and handler.
// It takes cfg of type *Config and handler of type BFFService.
// It returns a pointer to a Service struct.
func NewSevice(cfg *Config, handler BFFService) (*Service, error) {
	s := &Service{
		cfg:     cfg,
		handler: handler,
	}

	populateDefaults(cfg)

	dispatcher := dispatch.NewRouterDispatcher(s, parse)

	dispatcher.Use(middleware.NewErrorHandlingMiddleware())
	dispatcher.Use(middleware.NewMetricsMiddleware("bff-deriv", skipMetrics))
	dispatcher.Use(reqmid.NewTrottlerMiddleware(cfg.MaxRequests))

	requestLimitsFunc, err := getRequestLimits(cfg.RateLimits)
	if err != nil {
		return nil, err
	}

	dispatcher.Use(reqmid.NewRateLimiterMiddleware(requestLimitsFunc))

	registry := channel.NewConnectionRegistry(
		channel.WithMaxFrameLimit(maxMessageSize),
		channel.WithConcurrencyLimit(cfg.MaxRequestsPerConn),
	)
	endpoint := channel.NewChannel("/", dispatcher, registry, channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())
	endpoint.Use(httpmid.NewClientIPMiddleware(httpmid.CloudFront))

	s.server = server.NewServer(cfg.Listen)
	s.server.AddChannel(endpoint)
	s.server.AddHandler("/livez", http.HandlerFunc(s.HealthCheck))

	return s, nil
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

// getRequestLimits is a helper function transform request limits mentioned in server configuration
// into wasabi request limits, so that we can use the RateLimiter middleware.
func getRequestLimits(rateLimitCfg RateLimits) (func(wasabi.Request) (string, time.Duration, uint64), error) {
	if len(rateLimitCfg.Groups) == 0 {
		return getDefaultRequestLimits(rateLimitCfg.General)
	}

	return getRateLimitForMethods(rateLimitCfg)
}

func getRateLimitForMethods(rateLimitCfg RateLimits) (func(wasabi.Request) (string, time.Duration, uint64), error) {
	groupRatesMap, err := buildGroupRateMap(rateLimitCfg.Groups)

	if err != nil {
		return nil, err
	}

	generalRateLimitFunc, err := getDefaultRequestLimits(rateLimitCfg.General)

	if err != nil {
		return nil, err
	}

	return func(r wasabi.Request) (string, time.Duration, uint64) {
		groupRate, exists := groupRatesMap[r.RoutingKey()]

		if !exists {
			return generalRateLimitFunc(r)
		}

		ip := getIPFromRequest(r)

		return groupRate.Name + ip, groupRate.Interval, uint64(groupRate.Limit)
	}, nil
}

func buildGroupRateMap(groups []GroupRateLimits) (groupRatesMapType, error) {
	groupRatesMap := make(groupRatesMapType)

	for _, group := range groups {
		for _, method := range group.Methods {
			if _, exists := groupRatesMap[method]; exists {
				return nil, fmt.Errorf("method '%s' is repeated in multiple groups", method)
			}

			duration, err := time.ParseDuration(group.Limits.Interval)

			if err != nil {
				return nil, err
			}

			groupRatesMap[method] = struct {
				Name     string
				Methods  []string
				Interval time.Duration
				Limit    uint
			}{
				Name:     group.Name,
				Interval: duration,
				Limit:    group.Limits.Limit,
				Methods:  group.Methods,
			}
		}
	}

	return groupRatesMap, nil
}

// getDefaultRequestLimits is a helper function transform request limits using the default config
// into wasabi request limits, so that we can use the RateLimiter middleware.
func getDefaultRequestLimits(generalRateLimit GeneralRateLimits) (func(wasabi.Request) (string, time.Duration, uint64), error) {
	duration, err := time.ParseDuration(generalRateLimit.Interval)
	if err != nil {
		return nil, err
	}

	limit := uint64(generalRateLimit.Limit)

	return func(r wasabi.Request) (string, time.Duration, uint64) {
		ip := getIPFromRequest(r)
		return ip, duration, limit
	}, nil
}

func getIPFromRequest(r wasabi.Request) string {
	if ip, ok := r.Context().Value(httpmid.ClientIP).(string); ok {
		return ip
	}

	return "nil"
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
