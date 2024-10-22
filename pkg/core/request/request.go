package request

import (
	"context"
	"encoding/json"

	"github.com/ksysoev/wasabi"
)

const (
	BinaryMessage string = "binary"
	TextMessage   string = "text"
)

type Request struct {
	ctx         context.Context
	Params      map[string]any `json:"params"`
	ID          *int           `json:"req_id"`
	Method      string         `json:"method"`
	PassThrough any            `json:"passthrough"`
	data        []byte
}

// NewRequest creates a new Request object based on the provided message type and data.
// It takes ctx of type context.Context, msgType of type string, and data of type []byte.
// It returns a pointer to a Request object. If the msgType is BinaryMessage or if the data
// cannot be unmarshaled into a Request object, it initializes the Request with the provided
// msgType and data. Otherwise, it unmarshals the data into a Request object and sets the context and data fields.
func NewRequest(ctx context.Context, msgType string, data []byte) *Request {
	if msgType == BinaryMessage {
		return &Request{
			ctx:    ctx,
			data:   data,
			Method: msgType,
		}
	}

	var req Request

	err := json.Unmarshal(data, &req)
	if err != nil || req.Method == "" {
		return &Request{
			ctx:    ctx,
			data:   data,
			Method: msgType,
		}
	}

	req.ctx = ctx
	req.data = data

	return &req
}

// Data returns the data stored in the Request as a byte slice.
// It returns a byte slice containing the data.
func (r *Request) Data() []byte {
	return r.data
}

// RoutingKey returns the routing key for the request.
// It returns a string which is the method of the request.
func (r *Request) RoutingKey() string {
	return r.Method
}

// Context returns the context associated with the Request.
// It returns a context.Context which is the context stored in the Request.
func (r *Request) Context() context.Context {
	return r.ctx
}

// WithContext sets the context for the Request and returns the updated Request.
// It takes ctx of type context.Context.
// It returns the updated Request with the new context.
func (r *Request) WithContext(ctx context.Context) wasabi.Request {
	r.ctx = ctx
	return r
}
