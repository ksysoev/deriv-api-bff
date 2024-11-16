package processor

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type errorData struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}

// NewAPIError creates a new API error from the provided data.
// It takes a single parameter data of type any, which is expected to be a map with keys "code", "message", and optionally "details".
// It returns an error which is an instance of core.APIError if the data is in the expected format, or a descriptive error if the data format is incorrect.
// It returns an error if the data is not a map, if the "code" or "message" fields are missing or not strings, or if the "details" field cannot be marshaled to JSON.
func NewAPIError(data json.RawMessage) error {
	var errData errorData

	if err := json.Unmarshal(data, &errData); err != nil {
		return fmt.Errorf("failed to unmarshal APIError data: %w", err)
	}

	if errData.Code == "" || errData.Message == "" {
		return fmt.Errorf("error message is not in expected format")
	}

	return core.NewAPIError(errData.Code, errData.Message, errData.Details)
}
