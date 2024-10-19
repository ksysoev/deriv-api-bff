package composer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeParser() func([]byte) (map[string]any, error) {
	return func(data []byte) (map[string]any, error) {
		var res map[string]any

		if err := json.Unmarshal(data, &res); err != nil {
			return nil, err
		}

		return res, nil
	}
}

func TestComposer_WaitResponse_Success(t *testing.T) {
	composer := New()
	respChan := make(chan []byte, 1)
	parser := makeParser()
	ctx := context.Background()

	respChan <- []byte(`{"Params":"param1,param2","ReqID":1234}`)

	composer.Wait(ctx, "test", parser, respChan)

	resp, err := composer.Compose()

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "param1,param2", resp["Params"])
}

func TestComposer_WaitResponse_ParseError(t *testing.T) {
	composer := New()
	respChan := make(chan []byte, 1)
	respChan <- []byte("invalid json")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	composer.Wait(ctx, "test", makeParser(), respChan)

	_, err := composer.Compose()

	if !strings.HasPrefix(err.Error(), "fail to parse response: invalid character") {
		t.Errorf("expected error: %s, got something else: %s", err.Error(), err)
	}
}

func TestComposer_WaitResponse_ContextCancelled(t *testing.T) {
	composer := New()
	respChan := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	composer.Wait(ctx, "test", makeParser(), respChan)

	res, err := composer.Compose()
	assert.Nil(t, res)
	assert.Error(t, err)
}
