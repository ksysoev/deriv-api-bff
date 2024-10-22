package router

import (
	"context"
	"testing"

	core "github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mockDeriv := NewMockDerivAPI(t)
	mockHTTP := NewMockHTTPAPI(t)
	router := New(mockDeriv, mockHTTP)

	assert.NotNil(t, router)
	assert.Equal(t, mockDeriv, router.derivProv)
	assert.Equal(t, mockHTTP, router.httpProv)
}

func TestRouter_Handle_DerivProv(t *testing.T) {
	mockDeriv := NewMockDerivAPI(t)
	mockHTTP := NewMockHTTPAPI(t)
	router := New(mockDeriv, mockHTTP)
	mockConn := mocks.NewMockConnection(t)
	conn := core.NewConnection(mockConn, func(_ string) {})
	ctx := context.Background()
	mockRequest := request.NewRequest(ctx, request.TextMessage, []byte(`{"req_id":1,"method":"testMethod","params":{"key":"value"}}`))

	mockDeriv.EXPECT().Handle(conn, mockRequest).Return(nil)

	err := router.Handle(conn, mockRequest)

	assert.Nil(t, err)
}

func TestRouter_Handle_HTTPProv(t *testing.T) {
	mockDeriv := NewMockDerivAPI(t)
	mockHTTP := NewMockHTTPAPI(t)
	router := New(mockDeriv, mockHTTP)
	mockConn := mocks.NewMockConnection(t)
	conn := core.NewConnection(mockConn, func(_ string) {})
	ctx := context.Background()
	mockRequest := request.NewHTTPReq(ctx, "GET", "/test", nil)

	mockHTTP.EXPECT().Handle(conn, mockRequest).Return(nil)

	err := router.Handle(conn, mockRequest)

	assert.Nil(t, err)
}

func TestRouter_Handle_UnknownProv(t *testing.T) {
	mockDeriv := NewMockDerivAPI(t)
	mockHTTP := NewMockHTTPAPI(t)
	router := New(mockDeriv, mockHTTP)
	mockConn := mocks.NewMockConnection(t)
	conn := core.NewConnection(mockConn, func(_ string) {})
	mockRequest := mocks.NewMockRequest(t)

	err := router.Handle(conn, mockRequest)

	assert.Error(t, err)
}
