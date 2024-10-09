package core

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"testing"
	"time"
)

func initRequestProcessor() *RequestProcessor {
	tmpl, err := template.New("test").Parse("Params: {{.Params}}, ReqID: {{.ReqID}}")
	if err != nil {
		panic(fmt.Sprintf("failed to parse template: %v", err))
	}

	rp := &RequestProcessor{
		tempate: tmpl,
		allow:   []string{"Params", "ReqID"},
	}

	return rp
}

func TestComposer_WaitResponse_Success(t *testing.T) {
	composer := NewComposer()
	req := initRequestProcessor()
	respChan := make(chan []byte, 1)
	respChan <- []byte(`{"Params":"param1,param2","ReqID":1234}`)

	ctx := context.Background()
	reqID := 1234

	go composer.WaitResponse(ctx, req, respChan)

	resp, err := composer.Response(&reqID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"req_id":1234}`
	if string(resp) != expected {
		t.Fatalf("expected %s, got %s", expected, string(resp))
	}
}

func TestComposer_WaitResponse_ParseError(t *testing.T) {
	composer := NewComposer()
	req := initRequestProcessor()
	respChan := make(chan []byte, 1)
	respChan <- []byte("invalid json")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	composer.WaitResponse(ctx, req, respChan)

	_, err := composer.Response(nil)
	if !strings.HasPrefix(err.Error(), "fail to parse response : invalid character") {
		t.Fatalf("expected error: %s, got something else: %s", err.Error(), err)
	}
}

func TestComposer_WaitResponse_ContextCancelled(t *testing.T) {
	composer := NewComposer()
	req := initRequestProcessor()
	respChan := make(chan []byte, 1)
	respChan <- []byte(`{"Params":"param1,param2","ReqID":1234}`)

	var reqID = 1234

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	composer.WaitResponse(ctx, req, respChan)

	res, err := composer.Response(&reqID)
	if err == nil {
		t.Fatalf("expected error, got nil. While response was: %s", res)
	}
}
