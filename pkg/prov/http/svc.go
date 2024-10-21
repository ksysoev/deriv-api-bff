package http

import (
	"fmt"
	"net/http"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
)

type Service struct {
	handler wasabi.RequestHandler
}

func NewService() *Service {
	s := &Service{}

	s.handler = backend.NewBackend(s.requestFactory)

	return s
}

func (s *Service) requestFactory(r wasabi.Request) (*http.Request, error) {
	req, ok := r.(*core.Request)
	if !ok {
		return nil, fmt.Errorf("invalid request type %T", r)
	}

	url := fmt.Sprintf("http://localhost/%s", req.RoutingKey())

	httpReq, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return httpReq, nil
}
