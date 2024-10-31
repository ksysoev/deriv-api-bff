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

	req := NewHTTPReq(ctx, method, url, body, "testID")

	assert.Equal(t, "testID", req.ID())
	assert.Equal(t, ctx, req.ctx)
	assert.Equal(t, method, req.method)
	assert.Equal(t, url, req.url)
	assert.Equal(t, body, req.body)
	assert.NotNil(t, req.headers)
}

func TestAddHeader(t *testing.T) {
	req := NewHTTPReq(context.Background(), "GET", "http://example.com", nil, "testID")
	req.AddHeader("Content-Type", "application/json")

	assert.Equal(t, []string{"application/json"}, req.headers["Content-Type"])
}

func TestToHTTPRequest(t *testing.T) {
	ctx := context.Background()
	method := "POST"
	url := "http://example.com"
	body := []byte("test body")

	req := NewHTTPReq(ctx, method, url, body, "testID")
	req.AddHeader("Content-Type", "application/json")

	httpReq, err := req.ToHTTPRequest()
	assert.NoError(t, err)
	assert.Equal(t, method, httpReq.Method)
	assert.Equal(t, url, httpReq.URL.String())
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(httpReq.Body)
	assert.NoError(t, err)

	assert.Equal(t, body, buf.Bytes())
}

func TestToHTTPRequestError(t *testing.T) {
	ctx := context.Background()
	req := NewHTTPReq(ctx, "/invalid method", "http://example.com", nil, "testID")

	_, err := req.ToHTTPRequest()
	assert.Error(t, err)
}
func TestContext(t *testing.T) {
	ctx := context.Background()
	req := NewHTTPReq(ctx, "GET", "http://example.com", nil, "testID")

	assert.Equal(t, ctx, req.Context())
}

func TestRoutingKey(t *testing.T) {
	method := "GET"
	url := "http://example.com"
	req := NewHTTPReq(context.Background(), method, url, nil, "testID")

	expectedRoutingKey := method + " " + url
	assert.Equal(t, expectedRoutingKey, req.RoutingKey())
}

func TestData(t *testing.T) {
	body := []byte("test body")
	req := NewHTTPReq(context.Background(), "GET", "http://example.com", body, "testID")

	assert.Equal(t, body, req.Data())
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()

	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	req := NewHTTPReq(ctx, "GET", "http://example.com", nil, "testID")

	req1 := req.WithContext(newCtx)
	assert.Equal(t, newCtx, req1.Context())
}
