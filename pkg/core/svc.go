package core

import (
	"fmt"

	"github.com/ksysoev/wasabi"
)

type ConnRegistry interface {
	GetConnection(wasabi.Connection) *Conn
}

type DerivAPI interface {
	Handle(*Conn, *Request) error
}

type Service struct {
	be       DerivAPI
	ch       *CallHandler
	registry ConnRegistry
}

// NewService creates a new instance of Service.
// It takes cfg of type *Config, wsBackend of type DerivAPI, and connRegistry of type ConnRegistry.
// It returns a pointer to Service and an error.
// It returns an error if the call handler creation fails.
func NewService(cfg *Config, wsBackend DerivAPI, connRegistry ConnRegistry) (*Service, error) {
	callHandler, err := NewCallHandler(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to create call handler: %w", err)
	}

	return &Service{
		be:       wsBackend,
		ch:       callHandler,
		registry: connRegistry,
	}, nil
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

	iter, err := s.ch.Process(req)
	if err != nil {
		return err
	}

	if iter == nil {
		return fmt.Errorf("unsupported method: %s", req.RoutingKey())
	}

	ctx := req.Context()

	for ctx.Err() == nil && iter.HasNext() {
		data, err := iter.Next(conn.WaitResponse())
		if err == ErrIterDone {
			break
		}

		if err != nil {
			return err
		}

		r := &Request{data: data, Method: TextMessage, ctx: ctx}

		if err = s.be.Handle(conn, r); err != nil {
			return err
		}
	}

	resp, err := iter.WaitResp(req.ID)
	if err != nil {
		return err
	}

	return clientConn.Send(wasabi.MsgTypeText, resp)
}
