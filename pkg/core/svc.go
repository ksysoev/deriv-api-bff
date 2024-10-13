package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ksysoev/wasabi"
)

type Sender func(context.Context, []byte) error
type Waiter func() (reqID int64, respChan <-chan []byte)

type CallsRepo interface {
	GetCall(method string) Handler
}

type Handler interface {
	Handle(ctx context.Context, params map[string]any, watcher Waiter, send Sender) (map[string]any, error)
}

type ConnRegistry interface {
	GetConnection(wasabi.Connection) *Conn
}

type DerivAPI interface {
	Handle(*Conn, *Request) error
}

type Service struct {
	be       DerivAPI
	ch       CallsRepo
	registry ConnRegistry
}

// NewService creates a new instance of Service.
// It takes cfg of type *Config, wsBackend of type DerivAPI, and connRegistry of type ConnRegistry.
// It returns a pointer to Service and an error.
// It returns an error if the call handler creation fails.
func NewService(callRepo CallsRepo, wsBackend DerivAPI, connRegistry ConnRegistry) *Service {
	return &Service{
		be:       wsBackend,
		ch:       callRepo,
		registry: connRegistry,
	}
}

// PassThrough forwards a request to the backend service using the provided client connection.
// It takes clientConn of type wasabi.Connection and req of type *Request.
// It returns an error if the backend service fails to handle the request.
func (s *Service) PassThrough(clientConn wasabi.Connection, req *Request) error {
	conn := s.registry.GetConnection(clientConn)

	return s.be.Handle(conn, req)
}

// ProcessRequest processes a request by iterating over responses and handling them.
// It takes clientConn of type wasabi.Connection and req of type *Request.
// It returns an error if processing the request fails or if the method is unsupported.
func (s *Service) ProcessRequest(clientConn wasabi.Connection, req *Request) error {
	conn := s.registry.GetConnection(clientConn)

	handler := s.ch.GetCall(req.Method)

	if handler == nil {
		return fmt.Errorf("unsupported method: %s", req.Method)
	}

	resp, err := handler.Handle(
		req.Context(),
		req.Params,
		conn.WaitResponse,
		func(ctx context.Context, data []byte) error {
			return s.be.Handle(conn, &Request{
				Method: TextMessage,
				data:   data,
				ctx:    ctx,
			})
		},
	)

	if err != nil {
		return fmt.Errorf("failed to handle request: %w", err)
	}

	resp["req_id"] = req.ID

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return clientConn.Send(wasabi.MsgTypeText, data)
}
