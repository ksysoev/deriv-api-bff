package core

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	resp map[string]any
	call string
}

// NewAPIError creates a new instance of APIError with the provided call string and response data.
// It takes two parameters: call of type string and data of type map[string]any.
// It returns a pointer to an APIError struct.
func NewAPIError(call string, data map[string]any) *APIError {
	return &APIError{
		call: call,
		resp: data,
	}
}

// Error returns a formatted string describing the API error.
// It returns a string that includes the call information associated with the API error.
func (e *APIError) Error() string {
	return fmt.Sprintf("API error for call %s", e.call)
}

// Response generates a JSON-encoded response from the APIError instance.
// It takes req_id of type *int, which is an optional request identifier.
// It returns a byte slice containing the JSON-encoded response and an error if any.
// It returns an error if there is no response data available in the APIError instance.
// If req_id is nil, it returns the response without a request identifier.
// If req_id is provided, it includes the request identifier in the response.
func (e *APIError) Response(reqID *int) ([]byte, error) {
	if e.resp == nil {
		return nil, fmt.Errorf("no response data")
	}

	if reqID == nil {
		return json.Marshal(e.resp)
	}

	e.resp["req_id"] = *reqID

	return json.Marshal(e.resp)
}
