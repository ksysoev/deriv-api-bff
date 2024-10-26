package request

import (
	"context"
	"reflect"
	"testing"
)

func Test_NewRequest_TextData(t *testing.T) {
	ctx := context.Background()
	msgType := TextMessage
	data := []byte("cannot unmarshall this string")

	req := NewRequest(ctx, msgType, data)

	outputData := req.Data()

	if !reflect.DeepEqual(outputData, data) {
		t.Errorf("Expected outputData: %b to be equal to data: %b", outputData, data)
	}

	if req.RoutingKey() != TextMessage {
		t.Errorf("Expected routing key to be: %s but found: %s", TextMessage, req.RoutingKey())
	}

	if req.Context() != ctx {
		t.Errorf("Expected context to be %v, but found to be: %v", ctx, req.Context())
	}

	newCtx := context.WithoutCancel(ctx)

	newReq := req.WithContext(newCtx)

	if newReq.Context() != newCtx {
		t.Errorf("Expect new context to be %v, but found to be: %v", newReq.Context(), newCtx)
	}
}

func Test_NewRequest_Unmarshal(t *testing.T) {
	ctx := context.Background()
	msgType := "unknown"
	data := []byte(`{"params":{"p1":"v1"},"req_id":1234,"method":"text"}`)

	req := NewRequest(ctx, msgType, data)

	outputData := req.Data()

	if !reflect.DeepEqual(outputData, data) {
		t.Errorf("Expect outputData: %b to be equal to data: %b", outputData, data)
	}

	if req.Method != TextMessage {
		t.Errorf("Expected `Method` to be text, but found it to be: %s", req.Method)
	}
}

func Test_NewRequest_UnmarshalEmptyMethod(t *testing.T) {
	ctx := context.Background()
	msgType := TextMessage
	data := []byte(`{"params":{"p1":"v1"},"req_id":1234}`)

	req := NewRequest(ctx, msgType, data)

	outputData := req.Data()

	if !reflect.DeepEqual(outputData, data) {
		t.Errorf("Expect outputData: %b to be equal to data: %b", outputData, data)
	}

	if req.Method != TextMessage {
		t.Errorf("Expected `Method` to be text, but found it to be: %s", req.Method)
	}
}

func Test_NewRequest_Binary(t *testing.T) {
	ctx := context.Background()
	msgType := BinaryMessage
	data := []byte("foobar")

	req := NewRequest(ctx, msgType, data)

	outputData := req.Data()

	if !reflect.DeepEqual(outputData, data) {
		t.Errorf("Expect outputData: %b to be equal to data: %b", outputData, data)
	}

	if req.Method != BinaryMessage {
		t.Errorf("Expected `Method` to be text, but found it to be: %s", req.Method)
	}
}
