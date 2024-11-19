package composer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/response"
	"github.com/stretchr/testify/assert"
)

func makeParser(t *testing.T) func([]byte) (*response.Response, error) {
	t.Helper()

	return func(data []byte) (*response.Response, error) {
		var res map[string]json.RawMessage

		if err := json.Unmarshal(data, &res); err != nil {
			return nil, err
		}

		return response.New(data, res), nil
	}
}

func makeWaiter(t *testing.T) (respChan chan []byte, waiter func() (string, <-chan []byte)) {
	t.Helper()

	respChan = make(chan []byte, 1)

	return respChan, func() (string, <-chan []byte) {
		return "1234", respChan
	}
}

func TestComposer_Success(t *testing.T) {
	respChan, waiter := makeWaiter(t)
	composer := New(make(map[string][]string), waiter)
	parser := makeParser(t)
	ctx := context.Background()

	respChan <- []byte(`{"Params":"param1,param2","ReqID":1234}`)

	_, _, err := composer.Prepare(ctx, "test", parser)
	assert.NoError(t, err)

	resp, err := composer.Compose()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, json.RawMessage(`"param1,param2"`), resp["Params"])
}

func TestComposer_Compose_ParseError(t *testing.T) {
	respChan, waiter := makeWaiter(t)
	composer := New(make(map[string][]string), waiter)
	respChan <- []byte("invalid json")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _, err := composer.Prepare(ctx, "test", makeParser(t))
	assert.NoError(t, err)

	_, err = composer.Compose()

	if !strings.HasPrefix(err.Error(), "fail to parse response: invalid character") {
		t.Errorf("expected error: %s, got something else: %s", err.Error(), err)
	}
}

func TestComposer_Compose_ContextCancelled(t *testing.T) {
	_, waiter := makeWaiter(t)
	composer := New(make(map[string][]string), waiter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := composer.Prepare(ctx, "test", makeParser(t))
	assert.NoError(t, err)

	res, err := composer.Compose()
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestComposer_Prepare_DependenciesError(t *testing.T) {
	composer := New(map[string][]string{"test": {"dep1"}}, nil)
	composer.setError("dep1", assert.AnError)

	ctx := context.Background()
	_, _, err := composer.Prepare(ctx, "test", makeParser(t))
	assert.ErrorIs(t, err, assert.AnError)
}

func TestComposer_ComposeDependencies_NoDependencies(t *testing.T) {
	composer := New(make(map[string][]string), nil)
	ctx := context.Background()

	res, err := composer.composeDependencies(ctx, "test")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res)
}

func TestComposer_ComposeDependencies_WithDependencies(t *testing.T) {
	depGraph := map[string][]string{
		"test": {"dep1", "dep2"},
	}
	composer := New(depGraph, nil)
	composer.rawResps["dep1"] = "result1"
	composer.rawResps["dep2"] = "result2"
	composer.doneRequest("dep1")
	composer.doneRequest("dep2")

	ctx := context.Background()

	res, err := composer.composeDependencies(ctx, "test")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "result1", res["dep1"])
	assert.Equal(t, "result2", res["dep2"])
}

func TestComposer_ComposeDependencies_ContextCancelled(t *testing.T) {
	depGraph := map[string][]string{
		"test": {"dep1", "dep2"},
	}
	composer := New(depGraph, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := composer.composeDependencies(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}

func TestComposer_ComposeDependencies_ErrorInDependency(t *testing.T) {
	depGraph := map[string][]string{
		"test": {"dep1", "dep2"},
	}

	composer := New(depGraph, nil)
	composer.setError("dep1", assert.AnError)

	ctx := context.Background()
	res, err := composer.composeDependencies(ctx, "test")
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, res)
}
func TestComposer_DoneRequest_NewRequest(t *testing.T) {
	composer := New(make(map[string][]string), nil)
	name := "test"

	composer.doneRequest(name)

	ch, ok := composer.req[name]
	assert.True(t, ok)
	select {
	case <-ch:
		// Channel should be closed
	default:
		t.Error("expected channel to be closed")
	}
}

func TestComposer_DoneRequest_ExistingRequest(t *testing.T) {
	composer := New(make(map[string][]string), nil)
	name := "test"

	// Add a request manually
	composer.req[name] = make(chan struct{})

	composer.doneRequest(name)

	ch, ok := composer.req[name]
	assert.True(t, ok)
	select {
	case <-ch:
		// Channel should be closed
	default:
		t.Error("expected channel to be closed")
	}
}

func TestComposer_SetError_FirstError(t *testing.T) {
	composer := New(make(map[string][]string), nil)
	name := "test"
	err := assert.AnError

	composer.setError(name, err)

	assert.Equal(t, err, composer.err)
	ch, ok := composer.req[name]
	assert.True(t, ok)
	select {
	case <-ch:
		// Channel should be closed
	default:
		t.Error("expected channel to be closed")
	}
}

func TestComposer_SetError_SubsequentError(t *testing.T) {
	composer := New(make(map[string][]string), nil)
	name := "test"
	firstErr := assert.AnError
	secondErr := fmt.Errorf("another error")

	composer.setError(name, firstErr)
	composer.setError(name, secondErr)

	assert.Equal(t, firstErr, composer.err)
	ch, ok := composer.req[name]
	assert.True(t, ok)
	select {
	case <-ch:
		// Channel should be closed
	default:
		t.Error("expected channel to be closed")
	}
}
