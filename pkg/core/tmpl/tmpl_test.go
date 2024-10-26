package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmplRaw := `{"key": "${value}"}`
	tmpl, err := New(tmplRaw)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
}

func TestNew_InvalidTemplate(t *testing.T) {
	tmplRaw := `{"key": "${value"`
	tmpl, err := New(tmplRaw)
	assert.Error(t, err)
	assert.Nil(t, tmpl)
}

func TestMust(t *testing.T) {
	tmplRaw := `{"key": "${value}"}`
	assert.NotPanics(t, func() {
		tmpl := Must(tmplRaw)
		assert.NotNil(t, tmpl)
	})
}

func TestMust_Panic(t *testing.T) {
	tmplRaw := `{"key": "${value"`
	assert.Panics(t, func() {
		Must(tmplRaw)
	})
}

func TestExecute(t *testing.T) {
	tmplRaw := `{"key":"${value}"}`
	tmpl, err := New(tmplRaw)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)

	data := map[string]interface{}{
		"value": "test",
	}
	result, err := tmpl.Execute(data)
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"test"}`, string(result))
}

func TestExecute_InvalidData(t *testing.T) {
	tmplRaw := `{"key": "${value}"}`
	tmpl, err := New(tmplRaw)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)

	data := map[string]interface{}{
		"invalid": "test",
	}
	result, err := tmpl.Execute(data)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecute_InvalidParams(t *testing.T) {
	tmplRaw := `{"key": "${value}"}`
	tmpl, err := New(tmplRaw)
	assert.NoError(t, err)

	data := map[string]interface{}{
		"value": make(chan int),
	}
	result, err := tmpl.Execute(data)
	assert.Error(t, err)
	assert.Nil(t, result)
}
