package middleware

import (
	"context"
	"net/http"
	"net/url"
)

type ContextKey int

var keyQuery ContextKey = 1

// NewQueryParamsMiddleware creates a middleware that extracts query parameters from the URL
// and stores them in the request context.
// It returns a function that takes an http.Handler and returns an http.Handler.
// This middleware allows subsequent handlers to access query parameters via the request context.
func NewQueryParamsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyQuery, query)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// QueryParamsFromContext retrieves URL query parameters from the given context.
// It takes a single parameter ctx of type context.Context.
// It returns url.Values containing the query parameters if present, or nil if the context is nil or does not contain query parameters.
func QueryParamsFromContext(ctx context.Context) url.Values {
	if ctx == nil {
		return nil
	}

	if query, ok := ctx.Value(keyQuery).(url.Values); ok {
		return query
	}

	return nil
}
