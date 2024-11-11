package core

import (
	"context"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/ksysoev/wasabi"
	"github.com/valyala/fastjson"
)

type Conn struct {
	clientConn wasabi.Connection
	requests   map[string]chan []byte
	onClose    func(string)
	mu         sync.Mutex
}

// NewConnection initializes a new Conn instance with the provided wasabi.Connection and onClose callback.
// It takes conn of type wasabi.Connection and onClose of type func(id string).
// It returns a pointer to a Conn struct.
// It panics if conn is nil.
func NewConnection(conn wasabi.Connection, onClose func(id string)) *Conn {
	if conn == nil {
		panic("conn is nil")
	}

	return &Conn{
		clientConn: conn,
		requests:   make(map[string]chan []byte),
		onClose:    onClose,
	}
}

// ID returns the unique identifier of the connection.
// It returns a string which is the ID of the client connection.
func (c *Conn) ID() string {
	return c.clientConn.ID()
}

// Context returns the context associated with the Conn instance.
// It returns a context.Context which is the context of the underlying client connection.
func (c *Conn) Context() context.Context {
	return c.clientConn.Context()
}

// WaitResponse waits for a response from the connection and returns a request ID and a channel to receive the response.
// It returns an int64 representing the request ID and a receive-only channel of type []byte for the response.
func (c *Conn) WaitResponse() (reqID string, respChan <-chan []byte) {
	reqID = uuid.New().String()

	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan []byte, 1)
	c.requests[reqID] = ch

	return reqID, ch
}

// Send sends a message of the specified type through the connection.
// It takes msgType of type wasabi.MessageType and msg of type []byte.
// It returns an error if the message cannot be unmarshaled into the expected format or if there is an issue sending the message.
// If the message type is binary, it sends the message directly.
// If the message contains a req_id, it handles the request-response mechanism by sending the message to the appropriate channel.
func (c *Conn) Send(msgType wasabi.MessageType, msg []byte) error {
	if msgType == wasabi.MsgTypeBinary {
		return c.clientConn.Send(msgType, msg)
	}

	var parser fastjson.Parser

	v, err := parser.ParseBytes(msg)
	if err != nil {
		return c.clientConn.Send(msgType, msg)
	}

	reqID := v.GetStringBytes("passthrough", "req_id")

	if reqID == nil {
		return c.clientConn.Send(msgType, msg)
	}

	if c.DoneRequest(string(reqID), msg) {
		return nil
	}

	return c.clientConn.Send(msgType, msg)
}

func (c *Conn) DoneRequest(reqID string, resp []byte) bool {
	c.mu.Lock()
	ch, ok := c.requests[reqID]
	delete(c.requests, reqID)
	c.mu.Unlock()

	if ok {
		respCopy := make([]byte, len(resp))
		copy(respCopy, resp)

		ch <- respCopy

		return true
	}

	return false
}

// Close terminates the connection with a given status code and reason.
// It takes a status of type websocket.StatusCode, a reason of type string, and an optional closingCtx of type context.Context.
// It returns an error if the connection closure fails.
func (c *Conn) Close(status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	c.onClose(c.ID())
	return c.clientConn.Close(status, reason, closingCtx...)
}
