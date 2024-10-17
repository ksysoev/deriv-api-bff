package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	validator := NewMockValidator(t)
	processors := []RenderParser{NewMockRenderParser(t)}
	compFactory := func() WaitComposer {
		return NewMockWaitComposer(t)
	}

	handler := New(validator, processors, compFactory)

	assert.NotNil(t, handler)
	assert.Equal(t, validator, handler.validator)
	assert.Equal(t, processors, handler.processors)
	assert.NotNil(t, handler.newComposer)
}

func TestHandle_Success(t *testing.T) {
	expectedParams := map[string]any{"key": "value"}

	validator := NewMockValidator(t)
	validator.EXPECT().Validate(expectedParams).Return(nil)

	renderParser := NewMockRenderParser(t)
	renderParser.EXPECT().Render(mock.Anything, mock.Anything, expectedParams).Return(nil)

	waitComposer := NewMockWaitComposer(t)
	waitComposer.EXPECT().Compose().Return(expectedParams, nil)
	waitComposer.EXPECT().Wait(mock.Anything, mock.Anything, mock.Anything)

	handler := New(validator, []RenderParser{renderParser}, func() WaitComposer {
		return waitComposer
	})

	ctx := context.Background()

	echoChan := make(chan []byte, 1)
	waiter := func() (int64, <-chan []byte) {
		return 1, echoChan
	}

	sender := func(_ context.Context, data []byte) error {
		echoChan <- data
		return nil
	}

	resp, err := handler.Handle(ctx, expectedParams, waiter, sender)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"key": "value"}, resp)
}

func TestHandle_ValidationError(t *testing.T) {
	expectedParams := map[string]any{"key": "value"}

	validator := NewMockValidator(t)
	validator.EXPECT().Validate(expectedParams).Return(assert.AnError)

	renderParser := NewMockRenderParser(t)
	waitComposer := NewMockWaitComposer(t)

	handler := New(validator, []RenderParser{renderParser}, func() WaitComposer {
		return waitComposer
	})

	ctx := context.Background()

	echoChan := make(chan []byte, 1)
	waiter := func() (int64, <-chan []byte) {
		return 1, echoChan
	}

	sender := func(_ context.Context, data []byte) error {
		echoChan <- data
		return nil
	}

	resp, err := handler.Handle(ctx, expectedParams, waiter, sender)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, resp)
}

func TestHandle_SendError(t *testing.T) {
	expectedParams := map[string]any{"key": "value"}

	validator := NewMockValidator(t)
	validator.EXPECT().Validate(expectedParams).Return(nil)

	renderParser := NewMockRenderParser(t)
	renderParser.EXPECT().Render(mock.Anything, mock.Anything, expectedParams).Return(nil)

	waitComposer := NewMockWaitComposer(t)
	waitComposer.EXPECT().Wait(mock.Anything, mock.Anything, mock.Anything)

	handler := New(validator, []RenderParser{renderParser, renderParser}, func() WaitComposer {
		return waitComposer
	})

	ctx := context.Background()

	echoChan := make(chan []byte, 1)
	waiter := func() (int64, <-chan []byte) {
		return 1, echoChan
	}

	sender := func(_ context.Context, _ []byte) error {
		return assert.AnError
	}

	resp, err := handler.Handle(ctx, expectedParams, waiter, sender)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, resp)
}

func TestHandle_CancelledContext(t *testing.T) {
	expectedParams := map[string]any{"key": "value"}

	validator := NewMockValidator(t)
	validator.EXPECT().Validate(expectedParams).Return(nil)

	renderParser := NewMockRenderParser(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	waitComposer := NewMockWaitComposer(t)
	waitComposer.EXPECT().Compose().Return(nil, ctx.Err())

	handler := New(validator, []RenderParser{renderParser}, func() WaitComposer {
		return waitComposer
	})

	echoChan := make(chan []byte, 1)
	waiter := func() (int64, <-chan []byte) {
		return 1, echoChan
	}

	sender := func(_ context.Context, _ []byte) error {
		return nil
	}

	resp, err := handler.Handle(ctx, expectedParams, waiter, sender)
	assert.ErrorIs(t, err, ctx.Err())
	assert.Nil(t, resp)
}
