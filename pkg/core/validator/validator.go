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

func New(cfg Config) (*FieldValidator, error) {
	for field, fieldConfig := range cfg {
		if fieldConfig.Type != "string" && fieldConfig.Type != "number" && fieldConfig.Type != "bool" {
			return nil, fmt.Errorf("unknown type %s for field %s", fieldConfig.Type, field)
		}
	}

	return &FieldValidator{fields: cfg}, nil
}

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
		return errValidation.ApiError()
	}

	return nil
}

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
