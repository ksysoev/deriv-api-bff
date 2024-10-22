package core

import (
	"encoding/json"
	"errors"
	"fmt"
)

func createResponse(req *Request, resp map[string]any, err error) ([]byte, error) {
	var apiErr *APIError

	if errors.As(err, &apiErr) {
		resp = make(map[string]any)
		resp["error"] = apiErr.Encode()
	} else if err != nil {
		return nil, fmt.Errorf("failed to handle request: %w", err)
	}

	if req.ID != nil {
		resp["req_id"] = *req.ID
	}

	if req.PassThrough != nil {
		resp["passthrough"] = req.PassThrough
	}

	resp["echo"] = json.RawMessage(req.Data())
	resp["msg_type"] = req.RoutingKey()

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return data, nil
}
