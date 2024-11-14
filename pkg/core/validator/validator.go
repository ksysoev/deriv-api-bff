package validator

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Config map[string]any

type schemaValidator interface {
	Validate(data any) error
}

type FieldSchema struct {
	Type string `mapstructure:"type" json:"type"`
}

type FieldValidator struct {
	jsonSchema schemaValidator
}

// New creates a new FieldValidator based on the provided configuration.
// It takes cfg of type Config, which maps field names to their configurations.
// It returns a pointer to a FieldValidator and an error.
// It returns an error if any field in the configuration has an unknown type.
func New(cfg *Config) (*FieldValidator, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	required := make([]string, 0, len(*cfg))

	for field := range *cfg {
		required = append(required, field)
	}

	schema := map[string]any{
		"type":                 "object",
		"properties":           cfg,
		"additionalProperties": false,
		"required":             required,
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
func (v *FieldValidator) Validate(data []byte) error {
	var p map[string]any

	if data != nil {
		// TODO: Is it possible to find library to validate JSON schema against byte slice?
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("failed to unmarshal params: %w", err)
		}
	} else {
		p = make(map[string]any)
	}

	err := v.jsonSchema.Validate(p)

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
