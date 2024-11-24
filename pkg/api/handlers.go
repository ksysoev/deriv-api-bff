package api

import (
	"fmt"
	"net/http"

	request "github.com/ksysoev/deriv-api-bff/pkg/core/request"
	wasabi "github.com/ksysoev/wasabi"
)

// Handle processes a request received on a connection and routes it based on the request type.
// It takes conn of type wasabi.Connection and r of type wasabi.Request.
// It returns an error if the request type is unsupported or if the request type is empty.
// If the request type is core.TextMessage or core.BinaryMessage, it passes the request through to the handler.
// For other request types, it processes the request using the handler.
func (s *Service) Handle(conn wasabi.Connection, r wasabi.Request) error {
	req, ok := r.(*request.Request)
	if !ok {
		return fmt.Errorf("unsupported request type: %T", req)
	}

	switch req.RoutingKey() {
	case request.TextMessage, request.BinaryMessage:
		return s.handler.PassThrough(conn, req)
	case "":
		return fmt.Errorf("empty request type: %v", req)
	default:
		return s.handler.ProcessRequest(conn, req)
	}
}

// HealthCheck handles HTTP requests for checking the health status of the service.
// It takes a ResponseWriter to write the HTTP response and a Request which represents the client's request.
// It returns an HTTP status code 200 (OK) and a plain text message "OK".
func (s *Service) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
