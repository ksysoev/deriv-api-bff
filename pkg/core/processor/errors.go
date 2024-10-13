package processor

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

// NewAPIError creates a new API error from the provided data.
// It takes a single parameter data of type any, which is expected to be a map with keys "code", "message", and optionally "details".
// It returns an error which is an instance of core.APIError if the data is in the expected format, or a descriptive error if the data format is incorrect.
// It returns an error if the data is not a map, if the "code" or "message" fields are missing or not strings, or if the "details" field cannot be marshaled to JSON.
func NewAPIError(data any) error {
	err, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("error data is not in expected format")
	}

	code, ok := err["code"].(string)
	if !ok {
		return fmt.Errorf("error code is not in expected format")
	}

	message, ok := err["message"].(string)
	if !ok {
		return fmt.Errorf("error message is not in expected format")
	}

	var detailsData json.RawMessage

	if details, ok := err["details"]; ok {
		var err error

		detailsData, err = json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal APIError details: %w", err)
		}
	}

	return core.NewAPIError(code, message, detailsData)
}
