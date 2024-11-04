package repo

import (
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type CallsRepository struct {
	calls map[string]core.Handler
	mu    sync.RWMutex
}

// NewCallsRepository creates and returns a new instance of CallsRepository.
// It initializes the calls map to store core.Handler instances keyed by string.
// It returns a pointer to the newly created CallsRepository.
func NewCallsRepository() *CallsRepository {
	r := &CallsRepository{
		calls: make(map[string]core.Handler),
	}

	return r
}

// GetCall retrieves a handler function associated with the given method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a core.Handler which is the handler function associated with the specified method.
func (r *CallsRepository) GetCall(method string) core.Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.calls[method]
}

// UpdateCalls updates the repository with a new set of call handlers.
// It takes a callsMap parameter which is a map where the key is a string and the value is a core.Handler.
// This function does not return any values.
func (r *CallsRepository) UpdateCalls(callsMap map[string]core.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls = callsMap
}
