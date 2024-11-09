package core

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
)

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
