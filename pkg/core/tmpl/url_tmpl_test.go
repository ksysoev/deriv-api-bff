package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURLTmpl(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		wantErr bool
	}{
		{
			name:    "valid template",
			tmpl:    "http://example.com/${param}",
			wantErr: false,
		},
		{
			name:    "invalid template",
			tmpl:    "http://example.com/${param",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewURLTmpl(tt.tmpl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURLTmpl_Execute(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		params  any
		want    string
		wantErr bool
	}{
		{
			name:   "valid execution",
			tmpl:   "http://example.com/${param}",
			params: map[string]string{"param": "value"},
			want:   "http://example.com/value",
		},
		{
			name:    "missing parameter",
			tmpl:    "http://example.com/${param}",
			params:  map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid parameter type",
			tmpl:    "http://example.com/${param.invalid.path}",
			params:  map[string]int{"param": 123},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewURLTmpl(tt.tmpl)
			assert.NoError(t, err)

			got, err := tmpl.Execute(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}
