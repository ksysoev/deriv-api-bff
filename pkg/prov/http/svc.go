package http

import (
	"fmt"
	"net/http"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
	"github.com/ksysoev/wasabi/channel"
)

type Service struct {
	handler wasabi.RequestHandler
}

// NewService initializes and returns a new instance of Service.
// It sets up the handler with a new backend using the requestFactory.
// It returns a pointer to the newly created Service instance.
func NewService() *Service {
	s := &Service{}

	s.handler = backend.NewBackend(s.requestFactory)

	return s
}

func (s *Service) Handle(conn *core.Conn, req *request.HTTPReq) error {
	connWrap := channel.NewConnectionWrapper(conn, channel.WithSendWrapper(sendWrapper(conn, req)))

	return s.handler.Handle(connWrap, req)
}

// requestFactory creates an HTTP request from a wasabi.Request.
// It takes a parameter r of type wasabi.Request.
// It returns a pointer to an http.Request and an error.
// It returns an error if the request type is invalid or if the HTTP request creation fails.
func (s *Service) requestFactory(r wasabi.Request) (*http.Request, error) {
	req, ok := r.(*request.HTTPReq)
	if !ok {
		return nil, fmt.Errorf("invalid request type %T", r)
	}

	httpReq, err := req.ToHTTPRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	return httpReq, nil
}

func sendWrapper(conn *core.Conn, req *request.HTTPReq) channel.SendWrapper {
	return func(_ wasabi.Connection, _ wasabi.MessageType, msg []byte) error {
		if ok := conn.DoneRequest(req.ID(), msg); !ok {
			return fmt.Errorf("request ID %s not found is cancelled", req.ID())
		}

		return nil
	}
}
