package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHeadersMiddleware(t *testing.T) {
	// Create a test handler that will check the context for headers
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := HeadersFromContext(r.Context())
		if headers == nil {
			t.Error("Expected headers in context, got nil")
		}

		expected := "application/json"
		if headers.Get("Content-Type") != expected {
			t.Errorf("Expected header 'Content-Type' to be '%s', got '%s'", expected, headers.Get("Content-Type"))
		}
	})

	// Wrap the test handler with the middleware
	middleware := NewHeadersMiddleware()
	handler := middleware(testHandler)

	// Create a test request with headers
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check the response
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestHeadersFromContext_NilContext(t *testing.T) {
	headers := HeadersFromContext(context.TODO())
	if headers != nil {
		t.Error("Expected nil headers from nil context, got non-nil")
	}
}

func TestHeadersFromContext_NoHeaders(t *testing.T) {
	ctx := context.Background()
	headers := HeadersFromContext(ctx)
	if headers != nil {
		t.Error("Expected nil headers from context without headers, got non-nil")
	}
}
