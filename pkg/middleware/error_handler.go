package middleware

import (
	"log/slog"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

func NewErrorHandlingMiddleware() func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			defer func() {
				if r := recover(); r != nil {
					slog.Error(
						"panic during request handling",
						slog.Any("error", r),
						slog.String("routing_key", req.RoutingKey()),
						slog.Any("request", req),
					)
				}
			}()

			err := next.Handle(conn, req)

			if err != nil {
				slog.Error(
					"failed to handle request",
					slog.Any("error", err),
					slog.String("routing_key", req.RoutingKey()),
					slog.Any("request", req),
				)
			}

			return nil
		})
	}
}
