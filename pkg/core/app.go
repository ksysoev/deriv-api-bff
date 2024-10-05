package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/coder/websocket"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"golang.org/x/sync/singleflight"
)

type BackendForFE struct {
	mu         sync.Mutex
	be         wasabi.RequestHandler
	group      singleflight.Group
	connection map[string]*ConnState
	requests   map[int64]chan []byte
	ch         *CallHandler
}

func NewBackendForFE(wsBackend wasabi.RequestHandler, callHandler *CallHandler) *BackendForFE {
	return &BackendForFE{
		be:         wsBackend,
		connection: make(map[string]*ConnState),
		requests:   make(map[int64]chan []byte),
		ch:         callHandler,
	}
}

func (b *BackendForFE) Handle(conn wasabi.Connection, req wasabi.Request) error {
	b.mu.Lock()
	connState, ok := b.connection[conn.ID()]
	if !ok {
		wrap := channel.NewConnectionWrapper(conn,
			channel.WithSendWrapper(b.ResponseHandler),
			channel.WithCloseWrapper(b.CloseHandler),
		)
		connState = NewConnState(wrap)
		b.connection[conn.ID()] = connState
	}
	b.mu.Unlock()

	switch req.RoutingKey() {
	case TextMessage, BinaryMessage:
		return b.be.Handle(connState.Conn, req)
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

		for ctx.Err() == nil {
			id := connState.NextID()

			data, respChan, err := iter.Next(id)
			if err == ErrIterDone {
				break
			}

			if err != nil {
				return err
			}

			b.requests[id] = respChan

			r := &Request{data: data, Method: TextMessage, ctx: ctx}

			if err = b.be.Handle(connState.Conn, r); err != nil {
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

func (b *BackendForFE) ResponseHandler(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
	if msgType == wasabi.MsgTypeBinary {
		return conn.Send(msgType, msg)
	}

	var respID struct {
		ReqID int64 `json:"req_id"`
	}

	if err := json.Unmarshal(msg, &respID); err != nil {
		return conn.Send(msgType, msg)
	}

	if respID.ReqID == 0 {
		return conn.Send(msgType, msg)
	}

	b.mu.Lock()
	ch, ok := b.requests[respID.ReqID]
	b.mu.Unlock()
	if !ok {
		return conn.Send(msgType, msg)
	}

	buffer := make([]byte, len(msg))
	copy(buffer, msg)

	ch <- buffer

	return nil
}

func (b *BackendForFE) CloseHandler(conn wasabi.Connection, status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	b.mu.Lock()
	delete(b.connection, conn.ID())
	b.mu.Unlock()

	return conn.Close(status, reason, closingCtx...)
}
