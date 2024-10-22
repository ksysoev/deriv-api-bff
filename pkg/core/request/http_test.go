package request

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPReq(t *testing.T) {
	ctx := context.Background()
	method := "GET"
	url := "http://example.com"
	body := []byte("test body")

	req := NewHTTPReq(ctx, method, url, body)

	assert.Equal(t, ctx, req.ctx)
	assert.Equal(t, method, req.method)
	assert.Equal(t, url, req.url)
	assert.Equal(t, body, req.body)
	assert.NotNil(t, req.headers)
}

func TestAddHeader(t *testing.T) {
	req := NewHTTPReq(context.Background(), "GET", "http://example.com", nil)
	req.AddHeader("Content-Type", "application/json")

	assert.Equal(t, []string{"application/json"}, req.headers["Content-Type"])
}

func TestToHTTPRequest(t *testing.T) {
	ctx := context.Background()
	method := "POST"
	url := "http://example.com"
	body := []byte("test body")

	req := NewHTTPReq(ctx, method, url, body)
	req.AddHeader("Content-Type", "application/json")

	httpReq, err := req.ToHTTPRequest()
	assert.NoError(t, err)
	assert.Equal(t, method, httpReq.Method)
	assert.Equal(t, url, httpReq.URL.String())
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))

	buf := new(bytes.Buffer)
	buf.ReadFrom(httpReq.Body)
	assert.Equal(t, body, buf.Bytes())
}

func TestToHTTPRequestError(t *testing.T) {
	ctx := context.Background()
	req := NewHTTPReq(ctx, "/invalid method", "http://example.com", nil)

	_, err := req.ToHTTPRequest()
	assert.Error(t, err)
}
func TestContext(t *testing.T) {
	ctx := context.Background()
	req := NewHTTPReq(ctx, "GET", "http://example.com", nil)

	assert.Equal(t, ctx, req.Context())
}

func TestRoutingKey(t *testing.T) {
	method := "GET"
	url := "http://example.com"
	req := NewHTTPReq(context.Background(), method, url, nil)

	expectedRoutingKey := method + " " + url
	assert.Equal(t, expectedRoutingKey, req.RoutingKey())
}

func TestData(t *testing.T) {
	body := []byte("test body")
	req := NewHTTPReq(context.Background(), "GET", "http://example.com", body)

	assert.Equal(t, body, req.Data())
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	newCtx := context.WithValue(ctx, "key", "value")
	req := NewHTTPReq(ctx, "GET", "http://example.com", nil)

	req = req.WithContext(newCtx).(*HTTPReq)
	assert.Equal(t, newCtx, req.Context())
}
