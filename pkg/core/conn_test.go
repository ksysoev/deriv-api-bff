package core

import (
	"context"
	"testing"

	"github.com/coder/websocket"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewConnection(t *testing.T) {
	assert.Panics(t, func() {
		NewConnection(nil, func(_ string) {})
	})

	mockConn := mocks.NewMockConnection(t)
	onCloseCalled := false
	onClose := func(_ string) {
		onCloseCalled = true
	}

	conn := NewConnection(mockConn, onClose)

	assert.NotNil(t, conn)
	assert.Equal(t, mockConn, conn.clientConn)
	assert.Equal(t, initID, conn.currID)
	assert.NotNil(t, conn.requests)
	assert.Equal(t, 0, len(conn.requests))
	assert.NotNil(t, conn.onClose)
	assert.False(t, onCloseCalled)
}

func TestConn_ID(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	expectedID := "test-connection-id"
	mockConn.EXPECT().ID().Return(expectedID)

	conn := NewConnection(mockConn, func(_ string) {})

	actualID := conn.ID()

	assert.Equal(t, expectedID, actualID)
	mockConn.AssertExpectations(t)
}
func TestConn_Context(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	expectedCtx := context.Background()
	mockConn.EXPECT().Context().Return(expectedCtx)

	conn := NewConnection(mockConn, func(_ string) {})

	actualCtx := conn.Context()

	assert.Equal(t, expectedCtx, actualCtx)
}

func TestConn_WaitResponse(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	conn := NewConnection(mockConn, func(_ string) {})

	reqID, respChan := conn.WaitResponse()

	assert.NotNil(t, respChan)
	assert.Equal(t, initID+1, reqID)
	assert.Contains(t, conn.requests, reqID)
}

func TestConn_Close(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	expectedID := "test-connection-id"
	mockConn.EXPECT().ID().Return(expectedID)
	mockConn.EXPECT().Close(websocket.StatusNormalClosure, "normal closure").Return(nil)

	onCloseCalled := false
	onClose := func(id string) {
		onCloseCalled = true

		assert.Equal(t, expectedID, id)
	}

	conn := NewConnection(mockConn, onClose)

	err := conn.Close(websocket.StatusNormalClosure, "normal closure")

	assert.NoError(t, err)
	assert.True(t, onCloseCalled)
	mockConn.AssertExpectations(t)
}
func TestConn_Send(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)

	t.Run("Send binary message", func(t *testing.T) {
		msgType := wasabi.MsgTypeBinary
		msg := []byte("binary message")
		mockConn.EXPECT().Send(msgType, msg).Return(nil)

		conn := NewConnection(mockConn, func(_ string) {})
		err := conn.Send(msgType, msg)

		assert.NoError(t, err)
		mockConn.AssertExpectations(t)
	})

	t.Run("Send non-binary message without req_id", func(t *testing.T) {
		msgType := wasabi.MsgTypeText
		msg := []byte(`{"data":"test"}`)
		mockConn.EXPECT().Send(msgType, msg).Return(nil)

		conn := NewConnection(mockConn, func(_ string) {})
		err := conn.Send(msgType, msg)

		assert.NoError(t, err)
		mockConn.AssertExpectations(t)
	})

	t.Run("Send non-binary message with req_id", func(t *testing.T) {
		msgType := wasabi.MsgTypeText
		reqID := initID + 1
		msg := []byte(`{"req_id":1000001,"data":"test"}`)
		respChan := make(chan []byte, 1)

		conn := NewConnection(mockConn, func(_ string) {})
		conn.requests[reqID] = respChan

		err := conn.Send(msgType, msg)

		assert.NoError(t, err)
		assert.Equal(t, msg, <-respChan)

		_, exists := conn.requests[reqID]
		assert.False(t, exists)
	})

	t.Run("Send non-binary message with req_id but no matching request", func(t *testing.T) {
		msgType := wasabi.MsgTypeText
		msg := []byte(`{"req_id":1000001,"data":"test"}`)
		mockConn.EXPECT().Send(msgType, msg).Return(nil)

		conn := NewConnection(mockConn, func(_ string) {})
		err := conn.Send(msgType, msg)

		assert.NoError(t, err)
		mockConn.AssertExpectations(t)
	})
}
