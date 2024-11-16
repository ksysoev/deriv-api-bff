package response

type Response struct {
	filtered map[string]any
	data     map[string]any
}

func New(data map[string]any, filtered map[string]any) *Response {
	return &Response{
		data:     data,
		filtered: filtered,
	}
}

func (r *Response) Body() map[string]any {
	return r.data
}

func (r *Response) Filtered() map[string]any {
	return r.filtered
}
