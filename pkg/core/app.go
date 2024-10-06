package core

import (
	"fmt"
	"sync"

	"github.com/ksysoev/wasabi"
	"golang.org/x/sync/singleflight"
)

type BackendForFE struct {
	mu         sync.Mutex
	be         wasabi.RequestHandler
	group      singleflight.Group
	connection map[string]*Conn
	requests   map[int64]chan []byte
	ch         *CallHandler
}

func NewBackendForFE(wsBackend wasabi.RequestHandler, callHandler *CallHandler) *BackendForFE {
	return &BackendForFE{
		be:         wsBackend,
		connection: make(map[string]*Conn),
		requests:   make(map[int64]chan []byte),
		ch:         callHandler,
	}
}

func (b *BackendForFE) Handle(conn wasabi.Connection, req wasabi.Request) error {
	b.mu.Lock()
	id := conn.ID()
	coreConn, ok := b.connection[id]
	if !ok {
		onClose := func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			delete(b.connection, id)
		}

		coreConn = NewConnection(conn, onClose)
		b.connection[conn.ID()] = coreConn
	}
	b.mu.Unlock()

	switch req.RoutingKey() {
	case TextMessage, BinaryMessage:
		return b.be.Handle(coreConn, req)
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
			data, err := iter.Next(coreConn.WaitResponse())
			if err == ErrIterDone {
				break
			}

			if err != nil {
				return err
			}

			r := &Request{data: data, Method: TextMessage, ctx: ctx}

			if err = b.be.Handle(coreConn, r); err != nil {
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
