package router

import (
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
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

// New creates a new instance of Router with the provided DerivAPI service.
// It takes derivProv of type DerivAPI.
// It returns a pointer to a Router instance.
func New(derivProv DerivAPI, httpProv HTTPAPI) *Router {
	return &Router{
		derivProv: derivProv,
		httpProv:  httpProv,
	}
}

// Handle processes a request and delegates it to the appropriate provider based on the request type.
// It takes conn of type *core.Conn and req of type core.Request.
// It returns an error if the request type is unsupported or if the underlying provider's Handle method returns an error.
func (r *Router) Handle(conn *core.Conn, req core.Request) error {
	switch t := req.(type) {
	case *request.Request:
		return r.derivProv.Handle(conn, t)
	case *request.HTTPReq:
		return r.httpProv.Handle(conn, t)
	default:
		return fmt.Errorf("unsupported request type %T", req)
	}
}
