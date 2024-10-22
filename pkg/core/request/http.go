package request

import (
	"bytes"
	"context"
	"net/http"

	"github.com/ksysoev/wasabi"
)

type HTTPReq struct {
	ctx     context.Context
	url     string
	method  string
	body    []byte
	headers map[string][]string
}

// NewHTTPReq creates a new HTTPReq instance with the specified method, URL, and body.
// It takes three parameters: ctx of type context.Context, method of type string, and url of type string.
// It also takes body of type []byte which represents the request payload.
// It returns a pointer to an HTTPReq struct initialized with the provided method, URL, and body.
func NewHTTPReq(ctx context.Context, method, url string, body []byte) *HTTPReq {
	return &HTTPReq{
		ctx:     ctx,
		url:     url,
		method:  method,
		body:    body,
		headers: make(map[string][]string),
	}
}

// AddHeader adds a header key-value pair to the HTTP request.
// It takes two parameters: key of type string and value of type string.
// It appends the value to the list of values associated with the key in the headers map.
func (r *HTTPReq) AddHeader(key, value string) {
	r.headers[key] = append(r.headers[key], value)
}

// ToHTTPRequest converts an HTTPReq struct to an *http.Request.
// It takes no parameters and uses the fields of the HTTPReq struct.
// It returns a pointer to an http.Request and an error.
// It returns an error if the http.NewRequestWithContext call fails, such as when the method or URL is invalid.
func (r *HTTPReq) ToHTTPRequest() (*http.Request, error) {
	bodyReader := bytes.NewReader(r.body)
	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, values := range r.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

// Context returns the context associated with the HTTPReq.
// It takes no parameters.
// It returns a context.Context which is the context stored in the HTTPReq.
func (r *HTTPReq) Context() context.Context {
	return r.ctx
}

// RoutingKey constructs and returns a routing key string for the HTTP request.
// It concatenates the HTTP method and URL of the request, separated by a space.
// It does not return an error.
func (r *HTTPReq) RoutingKey() string {
	return r.method + " " + r.url
}

// Data returns the body of the HTTP request as a byte slice.
// It takes no parameters.
// It returns a byte slice containing the body of the HTTP request.
func (r *HTTPReq) Data() []byte {
	return r.body
}

// WithContext sets the context for the HTTP request.
// It takes ctx of type context.Context and returns the modified HTTPReq.
// This function does not return an error.
func (r *HTTPReq) WithContext(ctx context.Context) wasabi.Request {
	r.ctx = ctx
	return r
}
