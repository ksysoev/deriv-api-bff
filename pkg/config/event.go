package config

import (
	"context"
	"sync"
)

type Event[T any] struct {
	mu          *sync.Mutex
	subscribers []EventHandler[T]
}

func NewEvent[T any]() *Event[T] {
	return &Event[T]{
		subscribers: make([]EventHandler[T], 0),
		mu:          new(sync.Mutex),
	}
}

func (ce *Event[T]) RegisterHandler(handler EventHandler[T]) {
	ce.subscribers = append(ce.subscribers, handler)
}

func (ce *Event[T]) Notify(ctx context.Context, data T) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	var wg sync.WaitGroup

	for _, s := range ce.subscribers {
		wg.Add(1)

		go func(exec EventHandler[T]) {
			defer wg.Done()
			exec(ctx, data)
		}(s)
	}

	wg.Wait()
}
