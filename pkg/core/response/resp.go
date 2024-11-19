package response

import (
	"encoding/json"
)

type Response struct {
	filtered map[string]json.RawMessage
	data     json.RawMessage
}

// New creates a new Response instance with the provided data and filtered map.
// It takes data of type json.RawMessage and filtered of type map[string]json.RawMessage.
// It returns a pointer to a Response struct.
func New(data json.RawMessage, filtered map[string]json.RawMessage) *Response {
	return &Response{
		data:     data,
		filtered: filtered,
	}
}

// Body returns the body of the Response as a json.RawMessage.
func (r *Response) Body() json.RawMessage {
	return r.data
}

// Filtered returns the filtered response as a map of strings to json.RawMessage.
func (r *Response) Filtered() map[string]json.RawMessage {
	return r.filtered
}
