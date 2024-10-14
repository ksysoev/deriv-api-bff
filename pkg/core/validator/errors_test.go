package validator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError()

	if err == nil {
		t.Fatal("Expected a new ValidationError instance, got nil")
	}

	if err.HasErrors() {
		t.Fatal("Expected no errors in a new ValidationError instance")
	}
}

func TestAddError(t *testing.T) {
	err := NewValidationError()
	err.AddError("field1", errors.New("error1"))

	if !err.HasErrors() {
		t.Fatal("Expected errors to be present")
	}

	if err.Error() != "message is not valid: field1: error1" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}

func TestError(t *testing.T) {
	err := NewValidationError()
	err.AddError("field1", errors.New("error1"))
	err.AddError("field2", errors.New("error2"))

	expected := "message is not valid: field1: error1, field2: error2"
	if err.Error() != expected {
		t.Fatalf("Expected error string: %s, got: %s", expected, err.Error())
	}
}

func TestHasErrors(t *testing.T) {
	err := NewValidationError()

	if err.HasErrors() {
		t.Fatal("Expected no errors")
	}

	err.AddError("field1", errors.New("error1"))

	if !err.HasErrors() {
		t.Fatal("Expected errors to be present")
	}
}

func TestAPIError(t *testing.T) {
	err := NewValidationError()
	err.AddError("field1", errors.New("error1"))
	err.AddError("field2", errors.New("error2"))

	apiErr := err.APIError()
	if apiErr == nil {
		t.Fatal("Expected an APIError, got nil")
	}

	coreErr, ok := apiErr.(*core.APIError)
	if !ok {
		t.Fatalf("Expected core.APIError, got %T", apiErr)
	}

	if coreErr.Code != "InputValidationFailed" {
		t.Fatalf("Expected error code 'InputValidationFailed', got '%s'", coreErr.Code)
	}

	expectedDetails := map[string]string{
		"field1": "error1",
		"field2": "error2",
	}

	var actualDetails map[string]string

	if err := json.Unmarshal(coreErr.Details, &actualDetails); err != nil {
		t.Fatalf("Failed to unmarshal APIError details: %v", err)
	}

	for field, expectedErr := range expectedDetails {
		if actualDetails[field] != expectedErr {
			t.Fatalf("Expected error for field '%s': '%s', got: '%s'", field, expectedErr, actualDetails[field])
		}
	}
}
