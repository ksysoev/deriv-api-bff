package core

import (
	"encoding/json"
	"testing"
)

func TestNewAPIError(t *testing.T) {
	call := "TestCall"
	data := map[string]any{"key": "value"}

	apiError := NewAPIError(call, data)

	if apiError.call != call {
		t.Errorf("Expected call to be %s, got %s", call, apiError.call)
	}

	if apiError.resp["key"] != "value" {
		t.Errorf("Expected resp to have key 'key' with value 'value', got %v", apiError.resp["key"])
	}
}

func TestAPIError_Error(t *testing.T) {
	call := "TestCall"
	apiError := NewAPIError(call, nil)

	expectedError := "API error for call TestCall"
	if apiError.Error() != expectedError {
		t.Errorf("Expected error message to be %s, got %s", expectedError, apiError.Error())
	}
}

func TestAPIError_Response_NoReqID(t *testing.T) {
	data := map[string]any{"key": "value"}
	apiError := NewAPIError("TestCall", data)

	response, err := apiError.Response(nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var respData map[string]any
	if err := json.Unmarshal(response, &respData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if respData["key"] != "value" {
		t.Errorf("Expected response to have key 'key' with value 'value', got %v", respData["key"])
	}
}

func TestAPIError_Response_WithReqID(t *testing.T) {
	data := map[string]any{"key": "value"}
	apiError := NewAPIError("TestCall", data)

	reqID := 123
	response, err := apiError.Response(&reqID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var respData map[string]any
	if err := json.Unmarshal(response, &respData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if respData["key"] != "value" {
		t.Errorf("Expected response to have key 'key' with value 'value', got %v", respData["key"])
	}

	if respData["req_id"] != float64(reqID) { // JSON numbers are float64
		t.Errorf("Expected response to have req_id %d, got %v", reqID, respData["req_id"])
	}
}

func TestAPIError_Response_NoData(t *testing.T) {
	apiError := NewAPIError("TestCall", nil)

	_, err := apiError.Response(nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "no response data"
	if err.Error() != expectedError {
		t.Errorf("Expected error message to be %s, got %s", expectedError, err.Error())
	}
}
