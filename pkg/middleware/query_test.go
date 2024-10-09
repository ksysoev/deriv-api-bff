package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewQueryParamsMiddleware(t *testing.T) {
	// Create a test handler that will check the context for query parameters
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := QueryParamsFromContext(r.Context())
		if queryParams == nil {
			t.Error("Expected query parameters in context, got nil")
		}

		expected := "value"
		if queryParams.Get("key") != expected {
			t.Errorf("Expected query parameter 'key' to be '%s', got '%s'", expected, queryParams.Get("key"))
		}
	})

	// Wrap the test handler with the middleware
	middleware := NewQueryParamsMiddleware()
	handler := middleware(testHandler)

	// Create a test request with query parameters
	req := httptest.NewRequest("GET", "http://example.com/?key=value", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check the response
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestQueryParamsFromContext_NilContext(t *testing.T) {
	queryParams := QueryParamsFromContext(context.TODO())
	if queryParams != nil {
		t.Error("Expected nil query parameters from nil context, got non-nil")
	}
}

func TestQueryParamsFromContext_NoQueryParams(t *testing.T) {
	ctx := context.Background()
	queryParams := QueryParamsFromContext(ctx)
	if queryParams != nil {
		t.Error("Expected nil query parameters from context without query params, got non-nil")
	}
}


func TestQueryParamsFromContext_WithQueryParams(t *testing.T) {
	// Create a context with query parameters
	queryParams := url.Values{}
	queryParams.Set("key", "value")
	ctx := context.WithValue(context.Background(), keyQuery, queryParams)

	// Retrieve query parameters from context
	retrievedQueryParams := QueryParamsFromContext(ctx)

	// Check if the retrieved query parameters match the expected values
	if retrievedQueryParams == nil {
		t.Fatal("Expected non-nil query parameters, got nil")
	}

	expected := "value"
	if retrievedQueryParams.Get("key") != expected {
		t.Errorf("Expected query parameter 'key' to be '%s', got '%s'", expected, retrievedQueryParams.Get("key"))
	}
}
