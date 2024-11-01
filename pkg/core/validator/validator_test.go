package validator

import (
	"testing"
)

func TestNewFieldValidator(t *testing.T) {
	tests := []struct {
		config  Config
		name    string
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: Config{
				"name":  &FieldSchema{Type: "string"},
				"age":   &FieldSchema{Type: "number"},
				"admin": &FieldSchema{Type: "boolean"},
			},
			wantErr: false,
		},
		{
			name: "Invalid configuration with unknown type",
			config: Config{
				"unknown": &FieldSchema{Type: "unknown"},
			},
			wantErr: true,
		},
		{
			name: "Fail to marshal schema",
			config: Config{
				"unknown": make(chan int),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFieldValidator_Validate(t *testing.T) {
	config := Config{
		"name":  &FieldSchema{Type: "string"},
		"age":   &FieldSchema{Type: "number"},
		"admin": &FieldSchema{Type: "boolean"},
	}

	validator, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create FieldValidator: %v", err)
	}

	tests := []struct {
		data    map[string]any
		name    string
		wantErr bool
	}{
		{
			name: "Valid data",
			data: map[string]any{
				"name":  "John",
				"age":   30.0,
				"admin": true,
			},
			wantErr: false,
		},
		{
			name: "Missing field",
			data: map[string]any{
				"name": "John",
			},
			wantErr: true,
		},
		{
			name: "Extra field",
			data: map[string]any{
				"name":  "John",
				"age":   30.0,
				"admin": true,
				"extra": "field",
			},
			wantErr: true,
		},
		{
			name: "Invalid type",
			data: map[string]any{
				"name":  "John",
				"age":   "thirty",
				"admin": true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
