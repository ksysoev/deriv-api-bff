package processor

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

// NewAPIError creates a new instance of APIError with the provided call string and response data.
// It takes two parameters: call of type string and data of type map[string]any.
// It returns a pointer to an APIError struct.
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
