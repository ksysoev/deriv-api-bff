package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestNewMetricsMiddleware(t *testing.T) {
	otel.SetMeterProvider(noop.NewMeterProvider())

	middleware := NewMetricsMiddleware("test_scope", nil)

	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	handler := middleware(mockHandler)

	conn := mocks.NewMockConnection(t)
	req := mocks.NewMockRequest(t)
	req.EXPECT().RoutingKey().Return("test_routing_key")
	req.EXPECT().Context().Return(context.Background())

	err := handler.Handle(conn, req)
	assert.NoError(t, err)
}

func TestNewMetricsMiddleware_Error(t *testing.T) {
	otel.SetMeterProvider(noop.NewMeterProvider())

	middleware := NewMetricsMiddleware("test_scope", nil)

	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		time.Sleep(10 * time.Millisecond)
		return assert.AnError
	})

	handler := middleware(mockHandler)

	conn := mocks.NewMockConnection(t)
	req := mocks.NewMockRequest(t)
	req.EXPECT().RoutingKey().Return("test_routing_key")
	req.EXPECT().Context().Return(context.Background())

	err := handler.Handle(conn, req)
	assert.Error(t, err)
}

func TestNewMetricsMiddleware_Skip(t *testing.T) {
	otel.SetMeterProvider(noop.NewMeterProvider())

	middleware := NewMetricsMiddleware("test_scope", func(_ wasabi.Request) bool { return true })

	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return assert.AnError
	})

	handler := middleware(mockHandler)

	conn := mocks.NewMockConnection(t)
	req := mocks.NewMockRequest(t)

	err := handler.Handle(conn, req)
	assert.ErrorIs(t, err, assert.AnError)
}
