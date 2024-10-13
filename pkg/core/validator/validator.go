package validator

import (
	"fmt"
)

type Config map[string]*FieldSchema

type FieldSchema struct {
	Type string `mapstructure:"type"`
}

type FieldValidator struct {
	fields Config
}

// New creates a new FieldValidator based on the provided configuration.
// It takes cfg of type Config, which maps field names to their configurations.
// It returns a pointer to a FieldValidator and an error.
// It returns an error if any field in the configuration has an unknown type.
func New(cfg Config) (*FieldValidator, error) {
	for field, fieldConfig := range cfg {
		if fieldConfig.Type != "string" && fieldConfig.Type != "number" && fieldConfig.Type != "bool" {
			return nil, fmt.Errorf("unknown type %s for field %s", fieldConfig.Type, field)
		}
	}

	return &FieldValidator{fields: cfg}, nil
}

// Validate checks the provided data against the field configurations in the FieldValidator.
// It takes a single parameter data of type map[string]any which represents the data to be validated.
// It returns an error if there are validation errors, including missing required fields or fields that are not allowed.
// If the data contains fields not defined in the validator or if required fields are missing, it adds these errors to the validation error.
func (v *FieldValidator) Validate(data map[string]any) error {
	errValidation := NewValidationError()

	for field, config := range v.fields {
		if value, ok := data[field]; ok {
			if err := v.validateField(config, value); err != nil {
				errValidation.AddError(field, err)
			}
		} else {
			errValidation.AddError(field, fmt.Errorf("field %s is missing", field))
		}
	}

	for field := range data {
		if _, ok := v.fields[field]; !ok {
			errValidation.AddError(field, fmt.Errorf("field %s is not allowed", field))
		}
	}

	if errValidation.HasErrors() {
		return errValidation.APIError()
	}

	return nil
}

// validateField checks if the provided value matches the expected type defined in the FieldSchema.
// It takes a config of type *FieldSchema and a value of type any.
// It returns an error if the value does not match the expected type specified in the config.
// Possible error conditions include:
// - The value is not of type string when config.Type is "string".
// - The value is not of type float64 when config.Type is "number".
// - The value is not of type bool when config.Type is "bool".
// - The config.Type is unknown.
func (v *FieldValidator) validateField(config *FieldSchema, value any) error {
	switch config.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s is not a string", value)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("field %s is not an int", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s is not a bool", value)
		}
	default:
		return fmt.Errorf("unknown field type %s", config.Type)
	}

	return nil
}
