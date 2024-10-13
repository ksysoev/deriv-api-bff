package composer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
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
	expectedData := `{"Params":"param1,param2","ReqID":1234}`
	composer := New()
	respChan := make(chan []byte, 1)
	parser := makeParser()
	ctx := context.Background()

	respChan <- []byte(expectedData)

	go composer.Wait(ctx, parser, respChan)

	resp, err := composer.Compose()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp["Params"] != "param1,param2" {
		t.Fatalf("expected param1,param2, got %s", resp["Params"])
	}
}

func TestComposer_WaitResponse_ParseError(t *testing.T) {
	composer := New()
	respChan := make(chan []byte, 1)
	respChan <- []byte("invalid json")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	composer.Wait(ctx, makeParser(), respChan)

	_, err := composer.Compose()
	if !strings.HasPrefix(err.Error(), "fail to parse response: invalid character") {
		t.Fatalf("expected error: %s, got something else: %s", err.Error(), err)
	}
}

func TestComposer_WaitResponse_ContextCancelled(t *testing.T) {
	composer := New()
	respChan := make(chan []byte, 1)
	respChan <- []byte(`{"Params":"param1,param2","ReqID":1234}`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	composer.Wait(ctx, makeParser(), respChan)

	res, err := composer.Compose()
	if err == nil {
		t.Fatalf("expected error, got nil. While response was: %s", res)
	}
}
