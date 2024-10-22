package router

import (
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/http"
	"github.com/ksysoev/wasabi"
)

type DerivAPI interface {
	Handle(*core.Conn, *request.Request) error
}

type HTTPAPI interface {
	Handle(*core.Conn, *request.HTTPReq) error
}

type Router struct {
	derivProv DerivAPI
	httpProv  HTTPAPI
}

func New(derivApi DerivAPI) *Router {
	return &Router{
		derivProv: derivApi,
		httpProv:  http.NewService(),
	}
}

// TODO introduce core interface for requests objects
func (r *Router) Handle(conn *core.Conn, req wasabi.Request) error {
	switch t := req.(type) {
	case *request.Request:
		return r.derivProv.Handle(conn, t)
	case *request.HTTPReq:
		return r.httpProv.Handle(conn, t)
	default:
		return fmt.Errorf("unsupported request type %T", req)
	}
}
