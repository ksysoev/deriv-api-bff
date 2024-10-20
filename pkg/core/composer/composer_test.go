package composer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeParser(t *testing.T) func([]byte) (map[string]any, map[string]any, error) {
	t.Helper()

	return func(data []byte) (map[string]any, map[string]any, error) {
		var res map[string]any

		if err := json.Unmarshal(data, &res); err != nil {
			return nil, nil, err
		}

		return res, res, nil
	}
}

func makeWaiter(t *testing.T) (respChan chan []byte, waiter func() (int64, <-chan []byte)) {
	t.Helper()

	respChan = make(chan []byte, 1)

	return respChan, func() (int64, <-chan []byte) {
		return 1234, respChan
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
	assert.Equal(t, "param1,param2", resp["Params"])
}

func TestComposer_WaitResponse_ParseError(t *testing.T) {
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

func TestComposer_WaitResponse_ContextCancelled(t *testing.T) {
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
