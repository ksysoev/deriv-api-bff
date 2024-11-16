package core

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
)

// createResponse constructs a response based on the provided request, response data, and error.
// It takes req of type *request.Request, respData of type map[string]any, and err of type error.
// It returns a byte slice containing the marshaled response and an error if any occurs during processing.
// It returns an error if the request handling fails or if the response marshaling fails.
// If err is of type *APIError, it includes the encoded error in the response.
// If req.ID is not nil, it includes the request ID in the response.
// If req.PassThrough is not nil, it includes the passthrough data in the response.
// The response includes an "echo" field containing the raw request data.
func createResponse(req *request.Request, respData map[string]any, err error) ([]byte, error) {
	var apiErr *APIError

	resp := make(map[string]any)

	switch {
	case errors.As(err, &apiErr):
		resp["error"] = apiErr.Encode()
		resp["msg_type"] = "error"
	case err != nil:
		return nil, fmt.Errorf("failed to handle request: %w", err)
	default:
		resp["msg_type"] = req.RoutingKey()
		resp[req.RoutingKey()] = respData
	}

	if req.ID != nil {
		resp["req_id"] = *req.ID
	}

	if req.PassThrough != nil {
		resp["passthrough"] = req.PassThrough
	}

	resp["echo"] = json.RawMessage(req.Data())

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return data, nil
}
