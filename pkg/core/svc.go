package core

import (
	"context"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
)

type Request interface {
	Context() context.Context
	RoutingKey() string
	Data() []byte
}

type Sender func(Request) error
type Waiter func() (reqID string, respChan <-chan []byte)

type CallsRepo interface {
	GetCall(method string) Handler
	UpdateCalls(map[string]Handler)
}

type Handler interface {
	Handle(ctx context.Context, params map[string]any, watcher Waiter, send Sender) (map[string]any, error)
}

type ConnRegistry interface {
	GetConnection(wasabi.Connection) *Conn
}

type APIProvider interface {
	Handle(*Conn, Request) error
}

type Service struct {
	be       APIProvider
	ch       CallsRepo
	registry ConnRegistry
}

// NewService creates a new instance of Service.
// It takes cfg of type *Config, wsBackend of type DerivAPI, and connRegistry of type ConnRegistry.
// It returns a pointer to Service and an error.
// It returns an error if the call handler creation fails.
func NewService(callRepo CallsRepo, wsBackend APIProvider, connRegistry ConnRegistry) *Service {
	return &Service{
		be:       wsBackend,
		ch:       callRepo,
		registry: connRegistry,
	}
}

// PassThrough forwards a request to the backend service using the provided client connection.
// It takes clientConn of type wasabi.Connection and req of type *Request.
// It returns an error if the backend service fails to handle the request.
func (s *Service) PassThrough(clientConn wasabi.Connection, req *request.Request) error {
	conn := s.registry.GetConnection(clientConn)

	return s.be.Handle(conn, req)
}

// ProcessRequest handles an incoming request by delegating it to the appropriate handler based on the request method.
// It takes a client connection of type wasabi.Connection and a request of type *Request.
// It returns an error if the request method is unsupported, if the handler fails to process the request, or if the response cannot be marshaled to JSON.
// If the handler returns an APIError, it encodes the error in the response.
func (s *Service) ProcessRequest(clientConn wasabi.Connection, req *request.Request) error {
	conn := s.registry.GetConnection(clientConn)

	handler := s.ch.GetCall(req.Method)

	if handler == nil {
		return fmt.Errorf("unsupported method: %s", req.Method)
	}

	resp, err := handler.Handle(
		req.Context(),
		req.Params,
		conn.WaitResponse,
		func(req Request) error {
			return s.be.Handle(conn, req)
		},
	)

	data, err := createResponse(req, resp, err)
	if err != nil {
		return err
	}

	return clientConn.Send(wasabi.MsgTypeText, data)
}

func (s *Service) UpdateHandlers(handlers map[string]Handler) {
	s.ch.UpdateCalls(handlers)
}
