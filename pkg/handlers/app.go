package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/coder/websocket"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"golang.org/x/sync/singleflight"
)

type BackendForFE struct {
	mu         sync.Mutex
	be         wasabi.RequestHandler
	group      singleflight.Group
	connection map[string]wasabi.Connection
	currentID  int
	requests   map[int]chan []byte
	ch         *CallHandler
}

func NewBackendForFE(wsBackend wasabi.RequestHandler, callHandler *CallHandler) *BackendForFE {
	return &BackendForFE{
		be:         wsBackend,
		connection: make(map[string]wasabi.Connection),
		requests:   make(map[int]chan []byte),
		ch:         callHandler,
	}
}

func (b *BackendForFE) Handle(conn wasabi.Connection, req wasabi.Request) error {
	b.mu.Lock()
	connWrap, ok := b.connection[conn.ID()]
	if !ok {
		connWrap = channel.NewConnectionWrapper(conn,
			channel.WithSendWrapper(b.ResponseHandler),
			channel.WithCloseWrapper(b.CloseHandler),
		)
		b.connection[conn.ID()] = connWrap
	}
	b.mu.Unlock()

	switch req.RoutingKey() {
	case "text", "binary":
		return b.be.Handle(connWrap, req)
	default:
		iter, err := b.ch.Process(req)
		if err != nil {
			return err
		}

		if iter == nil {
			return fmt.Errorf("unsupported method: %s", req.RoutingKey())
		}

		for {
			b.currentID++
			id := b.currentID

			data, respChan, err := iter.Next(id)

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			b.requests[id] = respChan

			r := dispatch.NewRawRequest(req.Context(), wasabi.MsgTypeText, data)

			if err = b.be.Handle(connWrap, r); err != nil {
				return err
			}
		}

		resp := iter.WaitResp()

		respBytes, err := json.Marshal(resp)

		return conn.Send(wasabi.MsgTypeText, respBytes)
	}
}

func (b *BackendForFE) ResponseHandler(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
	if msgType == wasabi.MsgTypeBinary {
		return conn.Send(msgType, msg)
	}

	var respID struct {
		ReqID int `json:"req_id"`
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
