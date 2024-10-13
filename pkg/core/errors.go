package core

import "encoding/json"

type APIError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}

func NewAPIError(code, message string, details json.RawMessage) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Encode() json.RawMessage {
	bytes, err := json.Marshal(e)
	if err != nil {
		panic("failed to marshal APIError: " + err.Error())
	}

	return bytes
}
