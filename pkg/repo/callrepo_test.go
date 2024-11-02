package repo

import (
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewCallsRepository(t *testing.T) {
	repo := NewCallsRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.calls)
	assert.Empty(t, repo.calls)
}

func TestCallsRepository_GetCall(t *testing.T) {
	repo := NewCallsRepository()

	handler := repo.GetCall("nonexistent_method")
	assert.Nil(t, handler)

	expectedHandler := core.NewMockHandler(t)
	repo.calls["existing_method"] = expectedHandler

	handler = repo.GetCall("existing_method")
	assert.NotNil(t, handler)
	assert.Equal(t, expectedHandler, handler)
}

func TestCallsRepository_UpdateCalls(t *testing.T) {
	repo := NewCallsRepository()

	newCallsMap := map[string]core.Handler{
		"method1": core.NewMockHandler(t),
		"method2": core.NewMockHandler(t),
	}

	repo.UpdateCalls(newCallsMap)

	assert.Equal(t, newCallsMap, repo.calls)

	emptyCallsMap := map[string]core.Handler{}

	repo.UpdateCalls(emptyCallsMap)

	assert.Equal(t, emptyCallsMap, repo.calls)
}
