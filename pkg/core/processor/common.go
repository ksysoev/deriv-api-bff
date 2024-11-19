package processor

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// prepareResp processes a byte slice representing a JSON response body and returns a map of JSON raw messages.
// It takes data of type []byte.
// It returns a map[string]json.RawMessage containing the parsed JSON data and an error if any occurs.
// It returns an error if the data is empty or if the JSON unmarshalling fails.
func prepareResp(data []byte) (map[string]json.RawMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("response body not found")
	}

	switch data[0] {
	case '{':
		var respBody map[string]json.RawMessage

		if err := json.Unmarshal(data, &respBody); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}

		return respBody, nil
	case '[':
		return map[string]json.RawMessage{"list": data}, nil
	default:
		return map[string]json.RawMessage{"value": data}, nil
	}
}

// filterResp filters the response map to include only allowed keys and maps them to new keys if specified.
// It takes resp of type map[string]json.RawMessage, allow of type []string, and fieldMap of type map[string]string.
// It returns a map[string]json.RawMessage containing only the allowed keys, optionally mapped to new keys.
func filterResp(resp map[string]json.RawMessage, allow []string, fieldMap map[string]string) map[string]json.RawMessage {
	filtered := make(map[string]json.RawMessage, len(allow))

	for _, key := range allow {
		if _, ok := resp[key]; !ok {
			slog.Warn("Response body does not contain expeted key", slog.String("key", key))
			continue
		}

		destKey := key

		if fieldMap != nil {
			if mappedKey, ok := fieldMap[key]; ok {
				destKey = mappedKey
			}
		}

		filtered[destKey] = resp[key]
	}

	return filtered
}
