package core

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
	ctx    context.Context
	data   []byte
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
	ID     *int           `json:"req_id"`
}

func NewRequest(ctx context.Context, msgType string, data []byte) *Request {
	if msgType == BinaryMessage {
		return &Request{
			ctx:    ctx,
			data:   data,
			Method: string(msgType),
		}
	}

	var req Request

	err := json.Unmarshal(data, &req)
	if err != nil || req.Method == "" {
		return &Request{
			ctx:    ctx,
			data:   data,
			Method: string(msgType),
		}
	}

	req.ctx = ctx
	req.data = data

	return &req
}

func (r *Request) Data() []byte {
	return r.data
}

func (r *Request) RoutingKey() string {
	return r.Method
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) WithContext(ctx context.Context) wasabi.Request {
	r.ctx = ctx
	return r
}
