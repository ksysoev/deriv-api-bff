package handler

import "fmt"

type ValidatorConfig struct {
	fields map[string]*FielConfig
}

type FielConfig struct {
	Type string `mapstructure:"type"`
}

type Validator struct {
	config *ValidatorConfig
}

func NewValidator(config *ValidatorConfig) *Validator {
	return &Validator{config: config}
}

func (v *Validator) Validate(data map[string]any) error {
	errValidation := NewValidationError()

	for field, config := range v.config.fields {
		if value, ok := data[field]; ok {
			if err := v.validateField(config, value); err != nil {
				errValidation.AddError(field, err)
			}
		} else {
			errValidation.AddError(field, fmt.Errorf("field %s is missing", field))
		}
	}

	for field := range data {
		if _, ok := v.config.fields[field]; !ok {
			errValidation.AddError(field, fmt.Errorf("field %s is not allowed", field))
		}
	}

	return nil
}

func (v *Validator) validateField(config *FielConfig, value any) error {
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
