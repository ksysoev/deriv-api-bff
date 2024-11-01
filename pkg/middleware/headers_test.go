package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHeadersMiddleware(t *testing.T) {
	// Create a test handler that will check the context for headers
	testHandler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
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
	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	res := w.Result()

	defer res.Body.Close()

	// Check the response
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestHeadersFromContext_NoHeaders(t *testing.T) {
	if HeadersFromContext(context.Background()) != nil {
		t.Error("Expected nil headers from context without headers, got non-nil")
	}
}

func TestHeadersFromContext_NilContext(t *testing.T) {
	//nolint:staticcheck // Test nil context
	if QueryParamsFromContext(nil) != nil {
		t.Error("Expected nil headers from context without headers, got non-nil")
	}
}
