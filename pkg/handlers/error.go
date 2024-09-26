package handlers

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	call string
	resp map[string]any
}

func NewAPIError(call string, data map[string]any) *APIError {
	return &APIError{
		call: call,
		resp: data,
	}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error for call %s", e.call)
}

func (e *APIError) Response(req_id *int) ([]byte, error) {
	if e.resp == nil {
		return nil, fmt.Errorf("no response data")
	}

	if req_id == nil {
		return json.Marshal(e.resp)
	}

	e.resp["req_id"] = *req_id

	return json.Marshal(e.resp)
}
