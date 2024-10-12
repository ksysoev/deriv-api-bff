package handler

import (
	"encoding/json"
	"fmt"
)

type respError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details"`
}

type ValidationError struct {
	errors map[string]error
}

func NewValidationError() *ValidationError {
	return &ValidationError{errors: make(map[string]error)}
}

func (e *ValidationError) Error() string {
	errStr := "message is not valid:"
	for field, err := range e.errors {
		errStr += fmt.Sprintf(" %s: %v,", field, err)
	}

	errStr = errStr[:len(errStr)-1]

	return errStr
}

func (e *ValidationError) AddError(field string, err error) {
	e.errors[field] = err
}

func (e *ValidationError) HasErrors() bool {
	return len(e.errors) > 0
}

func (v *ValidationError) ErrorResponse(method string) (json.RawMessage, error) {
	if !v.HasErrors() {
		return nil, fmt.Errorf("no errors to generate response")
	}

	respError := respError{
		Code:    "InputValidationFailed",
		Message: fmt.Sprintf("Input validation failed: %s", method),
		Details: make(map[string]string, len(v.errors)),
	}

	for field, err := range v.errors {
		respError.Details[field] = err.Error()
	}

	return json.Marshal(respError)
}
