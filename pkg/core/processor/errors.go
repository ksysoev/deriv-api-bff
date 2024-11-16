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

// NewAPIError creates a new API error from the provided JSON data.
// It takes a single parameter data of type json.RawMessage.
// It returns an error if the JSON data cannot be unmarshaled or if the error message is not in the expected format.
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
