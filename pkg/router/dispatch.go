package router

import (
	"context"
	"encoding/json"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type Request struct {
	ctx     context.Context
	rawData []byte
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	ID      *int           `json:"req_id"`
}

func (r *Request) Data() []byte {
	return r.rawData
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

func Dispatch(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
	if msgType == wasabi.MsgTypeBinary {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	var req Request

	if err := json.Unmarshal(data, &req); err != nil {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	if req.Method == "" {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	req.ctx = ctx
	req.rawData = data

	return &req
}
