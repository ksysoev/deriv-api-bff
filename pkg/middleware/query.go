package middleware

import (
	"context"
	"net/http"
	"net/url"
)

type queryParamsKey struct{}

func NewQueryParamsMiddlewarefunc() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			ctx := r.Context()
			ctx = context.WithValue(ctx, queryParamsKey{}, query)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func QueryParamsFromContext(ctx context.Context) url.Values {
	if ctx == nil {
		return nil
	}

	if query, ok := ctx.Value(queryParamsKey{}).(map[string][]string); ok {
		return query
	}

	return nil
}
