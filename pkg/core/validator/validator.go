package validator

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Config map[string]*FieldSchema

type FieldSchema struct {
	Type string `mapstructure:"type"`
}

type FieldValidator struct {
	jsonSchema *jsonschema.Schema
}

// New creates a new FieldValidator based on the provided configuration.
// It takes cfg of type Config, which maps field names to their configurations.
// It returns a pointer to a FieldValidator and an error.
// It returns an error if any field in the configuration has an unknown type.
func New(cfg Config) (*FieldValidator, error) {
	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           make(map[string]interface{}),
		"additionalProperties": false,
		"required":             make([]string, 0),
	}

	for field, fieldConfig := range cfg {
		schema["properties"].(map[string]interface{})[field] = map[string]interface{}{"type": fieldConfig.Type}

		schema["required"] = append(schema["required"].([]string), field)
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	val, err := jsonschema.CompileString("", string(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &FieldValidator{
		jsonSchema: val,
	}, nil
}

// Validate checks the provided data against the field configurations in the FieldValidator.
// It takes a single parameter data of type map[string]any which represents the data to be validated.
// It returns an error if there are validation errors, including missing required fields or fields that are not allowed.
// If the data contains fields not defined in the validator or if required fields are missing, it adds these errors to the validation error.
func (v *FieldValidator) Validate(data map[string]any) error {
	err := v.jsonSchema.Validate(data)

	var errValidation *jsonschema.ValidationError

	switch {
	case errors.As(err, &errValidation):
		e := NewValidationError()

		for _, v := range errValidation.Causes {
			e.AddError(
				fmt.Sprintf("params%s", v.InstanceLocation),
				fmt.Errorf("%s", v.Message),
			)
		}

		if e.HasErrors() {
			return e.APIError()
		}

		return err
	case err != nil:
		return fmt.Errorf("failed to validate data: %w", err)
	}

	return nil
}
