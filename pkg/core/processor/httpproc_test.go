package processor

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/tmpl"
	"github.com/stretchr/testify/assert"
)

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
		cfg     *Config
		name    string
		wantErr bool
	}{
		{
			name: "Valid configuration",
			cfg: &Config{
				Name:        "TestProcessor",
				Method:      "GET",
				URLTemplate: "/test/url",
				Tmplt:       map[string]any{"key": "value"},
				FieldMap:    map[string]string{"key1": "mappedKey1"},
				Allow:       []string{"key1", "key2"},
			},
			wantErr: false,
		},
		{
			name: "Req template marshal error",
			cfg: &Config{
				Name:        "TestProcessor",
				Method:      "GET",
				URLTemplate: "/test/url",
				Tmplt:       map[string]any{"key": make(chan int)},
				FieldMap:    map[string]string{"key1": "mappedKey1"},
				Allow:       []string{"key1", "key2"},
			},
			wantErr: true,
		},
		{
			name: "Req template parse error",
			cfg: &Config{
				Name:        "TestProcessor",
				Method:      "GET",
				URLTemplate: "/test/url",
				Tmplt:       map[string]any{"test": "${params}invalid"},
				FieldMap:    map[string]string{"key1": "mappedKey1"},
				Allow:       []string{"key1", "key2"},
			},
			wantErr: true,
		},
		{
			name: "URL template parse error",
			cfg: &Config{
				Name:        "TestProcessor",
				Method:      "GET",
				URLTemplate: "/test/url${params",
				Tmplt:       map[string]any{"test": "${params}"},
				FieldMap:    map[string]string{"key1": "mappedKey1"},
				Allow:       []string{"key1", "key2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := NewHTTP(tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, processor)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, processor)
			assert.NotNil(t, processor.tmpl)
			assert.Equal(t, tt.cfg.FieldMap, processor.fieldMap)
			assert.Equal(t, tt.cfg.Allow, processor.allow)
		})
	}
}

func TestHTTPProc_Render(t *testing.T) {
	tests := []struct {
		proc          *HTTPProc
		param         map[string]any
		deps          map[string]any
		name          string
		wantRoutinKey string
		wantBody      []byte
		reqID         string
		wantErr       bool
	}{
		{
			name: "Valid URL and body template",
			proc: &HTTPProc{
				name:        "TestProcessor",
				method:      "POST",
				urlTemplate: tmpl.MustNewURLTmpl("http://example.com/${req_id}"),
				tmpl:        tmpl.MustNewTmpl(`{"param": "${params.param}"}`),
			},
			reqID:         "123",
			param:         map[string]any{"param": "value"},
			deps:          map[string]any{},
			wantRoutinKey: "POST http://example.com/123",
			wantBody:      []byte(`{"param": "value"}`),
			wantErr:       false,
		},
		{
			name: "Valid URL template, no body template",
			proc: &HTTPProc{
				name:        "TestProcessor",
				method:      "GET",
				urlTemplate: tmpl.MustNewURLTmpl("http://example.com/${req_id}"),
				tmpl:        nil,
			},
			reqID:         "123",
			param:         map[string]any{"param": "value"},
			deps:          map[string]any{},
			wantRoutinKey: "GET http://example.com/123",
			wantBody:      nil,
			wantErr:       false,
		},
		{
			name: "Invalid URL template",
			proc: &HTTPProc{
				name:        "TestProcessor",
				method:      "GET",
				urlTemplate: tmpl.MustNewURLTmpl("http://example.com/${params.invalid_field}"),
				tmpl:        nil,
			},
			reqID:         "123",
			param:         map[string]any{"param": "value"},
			deps:          map[string]any{},
			wantRoutinKey: "",
			wantBody:      nil,
			wantErr:       true,
		},
		{
			name: "Invalid body template",
			proc: &HTTPProc{
				name:        "TestProcessor",
				method:      "POST",
				urlTemplate: tmpl.MustNewURLTmpl("http://example.com/${req_id}"),
				tmpl:        tmpl.MustNewTmpl(`{"param": "${invalid_field}"}`),
			},
			reqID:         "123",
			param:         map[string]any{"param": "value"},
			deps:          map[string]any{},
			wantRoutinKey: "POST http://example.com/123",
			wantBody:      nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReq, err := tt.proc.Render(context.Background(), tt.reqID, tt.param, tt.deps)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotReq)
			} else {
				assert.Equal(t, tt.wantRoutinKey, gotReq.RoutingKey())

				if tt.wantBody == nil {
					assert.Nil(t, gotReq.Data())
				} else {
					assert.Equal(t, tt.wantBody, gotReq.Data())
				}
			}
		})
	}
}

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
		{
			name:    "JSON object with error key",
			data:    []byte(`{"error": "some error"}`),
			want:    nil,
			wantErr: true,
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
