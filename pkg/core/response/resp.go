package response

import (
	"encoding/json"
)

type Response struct {
	filtered map[string]json.RawMessage
	data     json.RawMessage
}

func New(data json.RawMessage, filtered map[string]json.RawMessage) *Response {
	return &Response{
		data:     data,
		filtered: filtered,
	}
}

func (r *Response) Body() json.RawMessage {
	return r.data
}

func (r *Response) Filtered() map[string]json.RawMessage {
	return r.filtered
}
