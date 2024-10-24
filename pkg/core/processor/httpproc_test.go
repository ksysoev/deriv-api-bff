package processor

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPProc_parse(t *testing.T) {
	tests := []struct {
		want    map[string]any
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Valid JSON object",
			data:    []byte(`{"key1": "value1", "key2": 2}`),
			want:    map[string]any{"key1": "value1", "key2": 2.0},
			wantErr: false,
		},
		{
			name:    "Valid JSON array",
			data:    []byte(`[{"key1": "value1"}, {"key2": 2}]`),
			want:    map[string]any{"list": []any{map[string]any{"key1": "value1"}, map[string]any{"key2": 2.0}}},
			wantErr: false,
		},
		{
			name:    "Valid single JSON value",
			data:    []byte(`"singleValue"`),
			want:    map[string]any{"value": "singleValue"},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			data:    []byte(`{"key1": "value1", "key2": 2`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Unexpected format",
			data:    []byte(`123`),
			want:    map[string]any{"value": 123.0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HTTPProc{}
			got, err := p.parse(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHTTPProc_Parse(t *testing.T) {
	tests := []struct {
		fieldMap map[string]string
		wantResp map[string]any
		wantFilt map[string]any
		name     string
		data     []byte
		allow    []string
		wantErr  bool
	}{
		{
			name:     "Valid JSON object with allowed keys",
			data:     []byte(`{"key1": "value1", "key2": 2}`),
			allow:    []string{"key1", "key2"},
			fieldMap: nil,
			wantResp: map[string]any{"key1": "value1", "key2": 2.0},
			wantFilt: map[string]any{"key1": "value1", "key2": 2.0},
			wantErr:  false,
		},
		{
			name:     "Valid JSON object with field mapping",
			data:     []byte(`{"key1": "value1", "key2": 2}`),
			allow:    []string{"key1", "key2"},
			fieldMap: map[string]string{"key1": "mappedKey1"},
			wantResp: map[string]any{"key1": "value1", "key2": 2.0},
			wantFilt: map[string]any{"mappedKey1": "value1", "key2": 2.0},
			wantErr:  false,
		},
		{
			name:     "Valid JSON object with missing allowed keys",
			data:     []byte(`{"key1": "value1"}`),
			allow:    []string{"key1", "key2"},
			fieldMap: nil,
			wantResp: map[string]any{"key1": "value1"},
			wantFilt: map[string]any{"key1": "value1"},
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			data:     []byte(`{"key1": "value1", "key2": 2`),
			allow:    []string{"key1", "key2"},
			fieldMap: nil,
			wantResp: nil,
			wantFilt: nil,
			wantErr:  true,
		},
		{
			name:     "Unexpected format",
			data:     []byte(`123`),
			allow:    []string{"value"},
			fieldMap: nil,
			wantResp: map[string]any{"value": 123.0},
			wantFilt: map[string]any{"value": 123.0},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HTTPProc{
				allow:    tt.allow,
				fieldMap: tt.fieldMap,
			}
			gotResp, gotFilt, err := p.Parse(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotResp)
				assert.Nil(t, gotFilt)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, gotResp)
				assert.Equal(t, tt.wantFilt, gotFilt)
			}
		})
	}
}
func TestHTTPProc_Name(t *testing.T) {
	tests := []struct {
		name     string
		procName string
		want     string
	}{
		{
			name:     "Valid name",
			procName: "TestProcessor",
			want:     "TestProcessor",
		},
		{
			name:     "Empty name",
			procName: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HTTPProc{name: tt.procName}
			got := p.Name()
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestNewHTTP(t *testing.T) {
	tests := []struct {
		cfg  *Config
		want *HTTPProc
		name string
	}{
		{
			name: "Valid configuration",
			cfg: &Config{
				Name:        "TestProcessor",
				Method:      "GET",
				URLTemplate: "/test/url",
				Tmplt:       template.Must(template.New("test").Parse("test template")),
				FieldMap:    map[string]string{"key1": "mappedKey1"},
				Allow:       []string{"key1", "key2"},
			},
			want: &HTTPProc{
				name:        "TestProcessor",
				method:      "GET",
				urlTemplate: "/test/url",
				tmpl:        template.Must(template.New("test").Parse("test template")),
				fieldMap:    map[string]string{"key1": "mappedKey1"},
				allow:       []string{"key1", "key2"},
			},
		},
		{
			name: "Empty configuration",
			cfg:  &Config{},
			want: &HTTPProc{
				name:        "",
				method:      "",
				urlTemplate: "",
				tmpl:        nil,
				fieldMap:    nil,
				allow:       nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHTTP(tt.cfg)
			assert.Equal(t, tt.want.name, got.name)
			assert.Equal(t, tt.want.method, got.method)
			assert.Equal(t, tt.want.urlTemplate, got.urlTemplate)
			assert.Equal(t, tt.want.tmpl, got.tmpl)
			assert.Equal(t, tt.want.fieldMap, got.fieldMap)
			assert.Equal(t, tt.want.allow, got.allow)
		})
	}
}
