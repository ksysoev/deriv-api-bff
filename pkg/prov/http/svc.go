package http

import (
	"fmt"
	"net/http"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
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
