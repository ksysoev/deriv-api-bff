package config

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	event := NewEvent[int]()
	integerEventStream := make([]int, 0)
	handlerFunc := func(_ context.Context, v int) {
		integerEventStream = append(integerEventStream, v)
	}

	event.RegisterHandler(handlerFunc)

	ctx := context.Background()

	event.Notify(ctx, 1)
	event.Notify(ctx, 2)

	assert.Equal(t, len(integerEventStream), 2)
	assert.Equal(t, []int{1, 2}, integerEventStream)
}

func TestEvent_MultipleHandlers(t *testing.T) {
	event := NewEvent[bool]()

	var cancelCounter atomic.Uint32

	var surviveCounter atomic.Uint32

	event.RegisterHandler(func(ctx context.Context, _ bool) {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			cancelCounter.Add(1)
		default:
			surviveCounter.Add(1)
		}
	})

	event.RegisterHandler(func(ctx context.Context, _ bool) {
		time.Sleep(300 * time.Millisecond)
		select {
		case <-ctx.Done():
			cancelCounter.Add(1)
		default:
			surviveCounter.Add(1)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	event.Notify(ctx, true)

	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	event.Notify(ctx2, true)

	assert.Equal(t, 3, int(cancelCounter.Load()))
	assert.Equal(t, 1, int(surviveCounter.Load()))
}
