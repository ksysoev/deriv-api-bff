package middleware

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func NewRedMetricsMiddleware(name string) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	meter := otel.GetMeterProvider().Meter(name)
	timing, err := meter.Float64Histogram(
		"request_duration",
		metric.WithDescription("Request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			start := time.Now()

			err := next.Handle(conn, req)

			elapsed := time.Since(start)

			hasError := "f"
			if err != nil {
				hasError = "t"
			}

			timing.Record(
				req.Context(),
				elapsed.Seconds(),
				metric.WithAttributes(
					attribute.String("routing_key", req.RoutingKey()),
					attribute.String("has_error", hasError),
				),
			)
			return err
		})
	}
}
