package handlerfactory

import (
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/stretchr/testify/assert"
)

func TestCreateHandler(t *testing.T) {
	tests := []struct {
		setupFunc func()
		name      string
		call      Config
		wantErr   bool
	}{
		{
			name: "valid deriv handler creation",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:     "backend1",
						FieldMap: map[string]string{"field1": "value1"},
						Tmplt:    map[string]any{"key1": "value1"},
						Allow:    []string{"allow1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid http handler creation",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:        "backend1",
						FieldMap:    map[string]string{"field1": "value1"},
						URLTemplate: "http://localhost/",
						Method:      "GET",
						Tmplt:       map[string]any{"key1": "value1"},
						Allow:       []string{"allow1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid processor configuration",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:     "backend1",
						FieldMap: map[string]string{"field1": "value1"},
						Allow:    []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid validator config",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "invalidType"}},
				Backend: []*processor.Config{
					{
						Name:     "backend1",
						FieldMap: map[string]string{"field1": "value1"},
						Tmplt:    map[string]any{"key1": "value1"},
						Allow:    []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid request template",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:     "backend1",
						FieldMap: map[string]string{"field1": "value1"},
						Tmplt:    nil,
						Allow:    []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid url template",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:        "backend1",
						FieldMap:    map[string]string{"field1": "value1"},
						URLTemplate: "http://localhost/${invalid",
						Allow:       []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "circular dependency",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						Name:      "backend1",
						FieldMap:  map[string]string{"field1": "value1"},
						Tmplt:     map[string]any{"key1": "value1"},
						DependsOn: []string{"backend2"},
						Allow:     []string{"allow1"},
					},
					{
						Name:      "backend2",
						FieldMap:  map[string]string{"field2": "value2"},
						Tmplt:     map[string]any{"key1": "value1"},
						DependsOn: []string{"backend1"},
						Allow:     []string{"allow2"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing name in backend config",
			call: Config{
				Method: "testMethod",
				Params: &validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*processor.Config{
					{
						FieldMap:    map[string]string{"field1": "value1"},
						URLTemplate: "http://localhost/${params.param1}",
						Allow:       []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing method in backend config",
			call: Config{
				Method: "",
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing params in backend config",
			call: Config{
				Method: "testMethod",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, handler, err := New(tt.call)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, name, tt.call.Method)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestTopSortDFS(t *testing.T) {
	tests := []struct {
		name    string
		input   []*processor.Config
		want    []*processor.Config
		wantErr bool
	}{
		{
			name: "no dependencies",
			input: []*processor.Config{
				{Name: "response1"},
				{Name: "response2"},
			},
			want: []*processor.Config{
				{Name: "response1"},
				{Name: "response2"},
			},
			wantErr: false,
		},
		{
			name: "simple dependency",
			input: []*processor.Config{
				{Name: "response1", DependsOn: []string{"response2"}},
				{Name: "response2"},
			},
			want: []*processor.Config{
				{Name: "response2"},
				{Name: "response1", DependsOn: []string{"response2"}},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			input: []*processor.Config{
				{Name: "response1", DependsOn: []string{"response2"}},
				{Name: "response2", DependsOn: []string{"response1"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "complex dependency",
			input: []*processor.Config{
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
				{Name: "response3"},
			},
			want: []*processor.Config{
				{Name: "response3"},
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
			},
			wantErr: false,
		},
		{
			name: "complex cicular dependency",
			input: []*processor.Config{
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
				{Name: "response3", DependsOn: []string{"response2"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Duplcate names",
			input: []*processor.Config{
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
				{Name: "response3", DependsOn: []string{"response2"}},
				{Name: "response1", DependsOn: []string{"response3"}},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := topSortDFS(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCreateComposerFactory(t *testing.T) {
	tests := []struct {
		graph map[string][]string
		name  string
	}{
		{
			name:  "empty graph",
			graph: map[string][]string{},
		},
		{
			name: "simple graph",
			graph: map[string][]string{
				"node1": {"node2"},
				"node2": {},
			},
		},
		{
			name: "complex graph",
			graph: map[string][]string{
				"node1": {"node2", "node3"},
				"node2": {"node3"},
				"node3": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := createComposerFactory(tt.graph)
			waiter := func() (string, <-chan []byte) {
				return "", nil
			}
			composer := factory(waiter)

			assert.NotNil(t, composer)
		})
	}
}
