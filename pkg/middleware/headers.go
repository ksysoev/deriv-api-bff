package middleware

import (
	"context"
	"net/http"
)

var keyHeaders ContextKey = 2

// NewHeadersMiddleware creates a middleware that injects request headers into the request context.
// It returns a function that takes an http.Handler and returns an http.Handler.
// This middleware allows subsequent handlers to access the request headers from the context.
func NewHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := r.Header
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyHeaders, headers)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// HeadersFromContext retrieves HTTP headers from the given context.
// It takes a single parameter ctx of type context.Context.
// It returns an http.Header containing the headers if present in the context, or nil if the context is nil or does not contain headers.
func HeadersFromContext(ctx context.Context) http.Header {
	if ctx == nil {
		return nil
	}

	if headers, ok := ctx.Value(keyHeaders).(http.Header); ok {
		return headers
	}

	return nil
}
