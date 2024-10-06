package core

import (
	"fmt"
	"sync"

	"github.com/ksysoev/wasabi"
	"golang.org/x/sync/singleflight"
)

type ConnRegistry interface {
	GetConnection(wasabi.Connection) *Conn
}

type BackendForFE struct {
	mu         sync.Mutex
	be         wasabi.RequestHandler
	group      singleflight.Group
	connection map[string]*Conn
	requests   map[int64]chan []byte
	ch         *CallHandler
	registry   ConnRegistry
}

func NewBackendForFE(wsBackend wasabi.RequestHandler, callHandler *CallHandler, connRegistry ConnRegistry) *BackendForFE {
	return &BackendForFE{
		be:       wsBackend,
		requests: make(map[int64]chan []byte),
		ch:       callHandler,
		registry: connRegistry,
	}
}

func (b *BackendForFE) Handle(clientConn wasabi.Connection, req wasabi.Request) error {
	conn := b.registry.GetConnection(clientConn)

	switch req.RoutingKey() {
	case TextMessage, BinaryMessage:
		return b.be.Handle(conn, req)
	default:
		r, ok := req.(*Request)
		if !ok {
			return fmt.Errorf("unsupported request type: %T", req)
		}

		iter, err := b.ch.Process(r)
		if err != nil {
			return err
		}

		if iter == nil {
			return fmt.Errorf("unsupported method: %s", req.RoutingKey())
		}

		ctx := r.Context()

		for ctx.Err() == nil && iter.HasNext() {
			data, err := iter.Next(conn.WaitResponse())
			if err == ErrIterDone {
				break
			}

			if err != nil {
				return err
			}

			r := &Request{data: data, Method: TextMessage, ctx: ctx}

			if err = b.be.Handle(conn, r); err != nil {
				return err
			}
		}

		resp, err := iter.WaitResp(r.ID)
		if err != nil {
			return err
		}

		return conn.Send(wasabi.MsgTypeText, resp)
	}
}
