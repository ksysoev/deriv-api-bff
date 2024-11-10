package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStrTmpl(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		expectError bool
	}{
		{
			name:        "Valid template",
			tmpl:        "Hello, ${name}!",
			expectError: false,
		},
		{
			name:        "Invalid template",
			tmpl:        "Hello, ${name!",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewStrTmpl(tt.tmpl)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)
			}
		})
	}
}

func TestMustNewStringTmpl(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		shouldPanic bool
	}{
		{
			name:        "Valid template",
			tmpl:        "Hello, ${name}!",
			shouldPanic: false,
		},
		{
			name:        "Invalid template",
			tmpl:        "Hello, ${name!",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					MustNewStringTmpl(tt.tmpl)
				})
			} else {
				assert.NotPanics(t, func() {
					tmpl := MustNewStringTmpl(tt.tmpl)
					assert.NotNil(t, tmpl)
				})
			}
		})
	}
}

func TestStrTmpl_Execute(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        string
		params      any
		expected    string
		expectError bool
	}{
		{
			name:        "Valid template with valid params",
			tmpl:        "Hello, ${name}!",
			params:      map[string]string{"name": "World"},
			expected:    "Hello, World!",
			expectError: false,
		},
		{
			name:        "Valid template with missing params",
			tmpl:        "Hello, ${name}!",
			params:      map[string]string{},
			expected:    "Hello, !",
			expectError: true,
		},
		{
			name:        "Valid template with integer params",
			tmpl:        "Hello, ${name}!",
			params:      map[string]int{"name": 123},
			expected:    "Hello, 123!",
			expectError: false,
		},
		{
			name:        "Invalid JSON params",
			tmpl:        "Hello, ${name}!",
			params:      func() {},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewStrTmpl(tt.tmpl)
			assert.NoError(t, err)
			assert.NotNil(t, tmpl)

			result, err := tmpl.Execute(tt.params)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
