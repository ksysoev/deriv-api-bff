package repo

import (
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type CallsRepository struct {
	mu    sync.RWMutex
	calls map[string]core.Handler
}

// NewCallsRepository creates a new instance of CallsRepository based on the provided configuration.
// It takes cfg of type *CallsConfig which contains the configuration for the calls.
// It returns a pointer to CallsRepository and an error if the repository creation fails.
// It returns an error if the validator creation fails or if there is an error parsing the request template.
func NewCallsRepository() *CallsRepository {
	r := &CallsRepository{
		calls: make(map[string]core.Handler),
	}

	return r
}

// GetCall retrieves a CallRunConfig based on the provided method name.
// It takes a single parameter method of type string, which specifies the method name.
// It returns a pointer to a CallRunConfig if the method exists in the repository, otherwise it returns nil.
func (r *CallsRepository) GetCall(method string) core.Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.calls[method]
}

// UpdateCalls refreshes the CallsConfig and rebuilds the handlers for calls accordingly
// It takes a single parameter that is pointer to the new calls config.
// The current implementation will completely overwrite the old config with new config.
// It returns an error to if any while building the handlers, otherwise it returns nil.
func (r *CallsRepository) UpdateCalls(callsMap map[string]core.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls = callsMap
}
