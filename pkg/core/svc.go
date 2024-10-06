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

func (s *Service) PassThrough(clientConn wasabi.Connection, req *Request) error {
	conn := s.registry.GetConnection(clientConn)

	return s.be.Handle(conn, req)
}

func (s *Service) ProcessReuest(clientConn wasabi.Connection, req *Request) error {
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
