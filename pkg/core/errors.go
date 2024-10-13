package core

import "encoding/json"

type APIError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}

// NewAPIError creates a new instance of APIError with the provided code, message, and details.
// It takes three parameters: code of type string, message of type string, and details of type json.RawMessage.
// It returns a pointer to an APIError struct populated with the provided values.
func NewAPIError(code, message string, details json.RawMessage) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error returns the message of the APIError.
// It returns a string containing the message of the APIError.
func (e *APIError) Error() string {
	return e.Message
}

// Encode serializes the APIError into a JSON-encoded byte slice.
// It takes no parameters.
// It returns a json.RawMessage containing the JSON-encoded representation of the APIError.
// It panics if the APIError cannot be marshaled into JSON.
func (e *APIError) Encode() json.RawMessage {
	bytes, err := json.Marshal(e)
	if err != nil {
		panic("failed to marshal APIError: " + err.Error())
	}

	return bytes
}
