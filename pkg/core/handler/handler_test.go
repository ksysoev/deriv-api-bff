package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
