package validator

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
)

type ValidationError struct {
	errors []fieldError
}

type fieldError struct {
	err   error
	field string
}

// NewValidationError creates and returns a new instance of ValidationError.
// It initializes the errors map to store validation errors.
// It returns a pointer to a ValidationError struct.
func NewValidationError() *ValidationError {
	return &ValidationError{errors: make([]fieldError, 0)}
}

// Error constructs a string representation of the ValidationError.
// It iterates over the errors map within the ValidationError struct,
// concatenating each field and its corresponding error message into a single string.
// It returns a string that describes all the validation errors.
// If there are no errors, it returns a string with the message "message is not valid".
func (e *ValidationError) Error() string {
	errStr := "message is not valid:"
	for _, fieldErr := range e.errors {
		errStr += fmt.Sprintf(" %s: %v,", fieldErr.field, fieldErr.err)
	}

	errStr = errStr[:len(errStr)-1]

	return errStr
}

// AddError adds an error to the ValidationError for a specific field.
// It takes field of type string and err of type error.
func (e *ValidationError) AddError(field string, err error) {
	e.errors = append(e.errors, fieldError{field: field, err: err})
}

// HasErrors checks if there are any validation errors present.
// It returns true if there is at least one error, otherwise false.
func (e *ValidationError) HasErrors() bool {
	return len(e.errors) > 0
}

// APIError converts a ValidationError into a core.APIError.
// It takes no parameters.
// It returns an error of type core.APIError with details about the validation errors.
func (e *ValidationError) APIError() error {
	details := make(map[string]string, len(e.errors))

	for _, fieldErr := range e.errors {
		details[fieldErr.field] = fieldErr.err.Error()
	}

	detailsData, err := json.Marshal(details)
	if err == nil {
		return core.NewAPIError("InputValidationFailed", "Input validation failed", detailsData)
	}

	return fmt.Errorf("failed to marshal APIError details: %w", err)
}
