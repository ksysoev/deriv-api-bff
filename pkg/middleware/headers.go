package middleware

import (
	"context"
	"net/http"
)

var keyHeaders ContextKey = 2

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

func HeadersFromContext(ctx context.Context) http.Header {
	if ctx == nil {
		return nil
	}

	if headers, ok := ctx.Value(keyHeaders).(http.Header); ok {
		return headers
	}

	return nil
}
