package middleware

import (
	"errors"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorHandlingMiddleware(t *testing.T) {
	tests := []struct {
		handlerErr  error
		name        string
		expectPanic bool
	}{
		{
			name:        "no error",
			handlerErr:  nil,
			expectPanic: false,
		},
		{
			name:        "handler error",
			handlerErr:  errors.New("handler error"),
			expectPanic: false,
		},
		{
			name:        "panic in handler",
			handlerErr:  nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewErrorHandlingMiddleware()
			handler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
				if tt.expectPanic {
					panic("panic in handler")
				}

				return tt.handlerErr
			})
			wrappedHandler := middleware(handler)

			conn := mocks.NewMockConnection(t)
			req := mocks.NewMockRequest(t)

			if tt.expectPanic || tt.handlerErr != nil {
				req.EXPECT().RoutingKey().Return("test-routing-key")
			}

			assert.NotPanics(t, func() {
				err := wrappedHandler.Handle(conn, req)
				assert.NoError(t, err)
			})
		})
	}
}
