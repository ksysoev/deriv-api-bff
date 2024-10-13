package validator

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

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

func (e *ValidationError) APIError() error {
	details := make(map[string]string, len(e.errors))

	for field, err := range e.errors {
		details[field] = err.Error()
	}

	detailsData, err := json.Marshal(details)
	if err != nil {
		panic("failed to marshal APIError details: " + err.Error())
	}

	return core.NewAPIError("InputValidationFailed", "Input validation failed", detailsData)
}
