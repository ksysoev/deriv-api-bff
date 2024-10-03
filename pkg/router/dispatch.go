package router

import (
	"context"

	"github.com/ksysoev/wasabi"
)

type Request struct {
	Ctx     context.Context
	RawData []byte
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	ID      *int           `json:"req_id"`
}

func (r *Request) Data() []byte {
	return r.RawData
}

func (r *Request) RoutingKey() string {
	return r.Method
}

func (r *Request) Context() context.Context {
	return r.Ctx
}

func (r *Request) WithContext(ctx context.Context) wasabi.Request {
	r.Ctx = ctx
	return r
}
