package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api/schema"
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
	case "new_ping":
		typ, resp, err := b.handleNewPingRequest(connWrap, req)
		if err != nil {
			return err
		}

		return conn.Send(typ, resp)
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

			data, err := iter.Next(id)

			if err != nil {
				return err
			}

			if data == nil {
				break
			}

			fmt.Println("data", string(data))

			r := dispatch.NewRawRequest(req.Context(), wasabi.MsgTypeText, data)
			if err = b.be.Handle(connWrap, r); err != nil {
				return err
			}
		}

		return nil
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

	ch <- msg
	return nil
}

func (b *BackendForFE) CloseHandler(conn wasabi.Connection, status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	b.mu.Lock()
	delete(b.connection, conn.ID())
	b.mu.Unlock()

	return conn.Close(status, reason, closingCtx...)
}

func (b *BackendForFE) handleNewPingRequest(conn wasabi.Connection, req wasabi.Request) (wasabi.MessageType, []byte, error) {

	b.mu.Lock()
	b.currentID++
	id1 := b.currentID
	b.currentID++
	id2 := b.currentID
	ch1 := make(chan []byte, 1)
	ch2 := make(chan []byte, 1)
	b.requests[id1] = ch1
	b.requests[id2] = ch2
	b.mu.Unlock()

	timeReq := schema.Time{
		Time:  1,
		ReqId: &id1,
	}

	timeReqBytes, err := json.Marshal(timeReq)
	if err != nil {
		return wasabi.MsgTypeText, nil, err
	}

	r := dispatch.NewRawRequest(req.Context(), wasabi.MsgTypeText, timeReqBytes)
	if err = b.be.Handle(conn, r); err != nil {
		return wasabi.MsgTypeText, nil, err
	}

	pingReq := schema.Ping{
		Ping:  1,
		ReqId: &id2,
	}

	pingReqBytes, err := json.Marshal(pingReq)
	if err != nil {
		return wasabi.MsgTypeText, nil, err
	}

	r = dispatch.NewRawRequest(req.Context(), wasabi.MsgTypeText, pingReqBytes)
	if err = b.be.Handle(conn, r); err != nil {
		return wasabi.MsgTypeText, nil, err
	}

	respBytes := make([]byte, 0)

	timeOutCh := time.After(5 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case resp := <-ch1:
			respBytes = append(respBytes, resp...)
			ch1 = nil
		case resp := <-ch2:
			respBytes = append(respBytes, resp...)
			ch2 = nil
		case <-timeOutCh:
			return wasabi.MsgTypeText, nil, fmt.Errorf("timeout")
		}
	}

	return wasabi.MsgTypeText, respBytes, nil
}
