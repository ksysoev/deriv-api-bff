package repo

import (
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/stretchr/testify/assert"
)

func TestNewCallsRepository(t *testing.T) {
	tests := []struct {
		cfg     *CallsConfig
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &CallsConfig{
				Calls: []CallConfig{
					{
						Method: "testMethod",
						Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: map[string]any{"key1": "value1"},
								Allow:           []string{"allow1"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid template",
			cfg: &CallsConfig{
				Calls: []CallConfig{
					{
						Method: "testMethod",
						Params: validator.Config{"param1": validator.FieldSchema{Type: "value1"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: map[string]any{"key1": "${value1"},
								Allow:           []string{"allow1"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "circular dependency",
			cfg: &CallsConfig{
				Calls: []CallConfig{
					{
						Method: "testMethod",
						Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: map[string]any{"key1": "value1"},
								DependsOn:       []string{"responseBody2"},
								Allow:           []string{"allow1"},
							},
							{
								FieldsMap:       map[string]string{"field2": "value2"},
								ResponseBody:    "responseBody2",
								RequestTemplate: map[string]any{"key1": "value1"},
								DependsOn:       []string{"responseBody1"},
								Allow:           []string{"allow2"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid dependency",
			cfg: &CallsConfig{
				Calls: []CallConfig{
					{
						Method: "testMethod",
						Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: map[string]any{"key1": "value1"},
								DependsOn:       []string{"responseBody2"},
								Allow:           []string{"allow1"},
							},
							{
								FieldsMap:       map[string]string{"field2": "value2"},
								ResponseBody:    "responseBody2",
								RequestTemplate: map[string]any{"key1": "value1"},
								Allow:           []string{"allow2"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCallsRepository(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Contains(t, got.calls, "testMethod")
			}
		})
	}
}

func TestGetCall(t *testing.T) {
	repo, err := NewCallsRepository(&CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	tests := []struct {
		name   string
		method string
		found  bool
	}{
		{
			name:   "existing method",
			method: "testMethod",
			found:  true,
		},
		{
			name:   "non-existing method",
			method: "nonExistingMethod",
			found:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callConfig := repo.GetCall(tt.method)

			if tt.found {
				assert.NotNil(t, callConfig)
			} else {
				assert.Nil(t, callConfig)
			}
		})
	}
}
func TestTopSortDFS(t *testing.T) {
	tests := []struct {
		name    string
		input   []*BackendConfig
		want    []*BackendConfig
		wantErr bool
	}{
		{
			name: "no dependencies",
			input: []*BackendConfig{
				{Name: "response1"},
				{Name: "response2"},
			},
			want: []*BackendConfig{
				{Name: "response1"},
				{Name: "response2"},
			},
			wantErr: false,
		},
		{
			name: "simple dependency",
			input: []*BackendConfig{
				{Name: "response1", DependsOn: []string{"response2"}},
				{Name: "response2"},
			},
			want: []*BackendConfig{
				{Name: "response2"},
				{Name: "response1", DependsOn: []string{"response2"}},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			input: []*BackendConfig{
				{Name: "response1", DependsOn: []string{"response2"}},
				{Name: "response2", DependsOn: []string{"response1"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "complex dependency",
			input: []*BackendConfig{
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
				{Name: "response3"},
			},
			want: []*BackendConfig{
				{Name: "response3"},
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
			},
			wantErr: false,
		},
		{
			name: "complex cicular dependency",
			input: []*BackendConfig{
				{Name: "response1", DependsOn: []string{"response3"}},
				{Name: "response2", DependsOn: []string{"response1"}},
				{Name: "response3", DependsOn: []string{"response2"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Duplcate names",
			input: []*BackendConfig{
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

func TestUpdateCalls_ExistingMethod_Success(t *testing.T) {
	oldCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
		},
	}

	callsRepo, err := NewCallsRepository(oldCallsConfig)
	if err != nil {
		t.Errorf("Unexpected Error: %v", err)
	}

	oldHandler := callsRepo.GetCall("testMethod")

	newCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						DependsOn:       []string{"responseBody2"},
						Allow:           []string{"allow1"},
					},
					{
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow2"},
					},
				},
			},
		},
	}

	callsRepo.UpdateCalls(newCallsConfig)

	newHandler := callsRepo.GetCall("testMethod")

	assert.NotEqualValues(t, oldHandler, newHandler, "old handler and new handler must be different")
}

func TestUpdateCalls_NewMethod_Success(t *testing.T) {
	oldCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
		},
	}

	callsRepo, err := NewCallsRepository(oldCallsConfig)
	if err != nil {
		t.Errorf("Unexpected Error: %v", err)
	}

	oldHandler := callsRepo.GetCall("testMethod")

	newCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethodNew",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						DependsOn:       []string{"responseBody2"},
						Allow:           []string{"allow1"},
					},
					{
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow2"},
					},
				},
			},
		},
	}

	callsRepo.UpdateCalls(newCallsConfig)

	newHandler := callsRepo.GetCall("testMethodNew")

	assert.Nil(t, callsRepo.GetCall("testMethod"), "testMethod handler does not exist anymore")
	assert.NotEqualValues(t, oldHandler, newHandler, "old handler and new handler must be different")
}

func TestUpdateCalls_Failure(t *testing.T) {
	oldCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
		},
	}

	callsRepo, err := NewCallsRepository(oldCallsConfig)
	if err != nil {
		t.Errorf("Unexpected Error: %v", err)
	}

	oldHandler := callsRepo.GetCall("testMethod")

	newCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "value1"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "${value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
		},
	}

	callsRepo.UpdateCalls(newCallsConfig)

	newHandler := callsRepo.GetCall("testMethod")
	assert.Equal(t, oldHandler, newHandler, "old handler and new handler must be same")
}

func TestCreateHandler(t *testing.T) {
	tests := []struct {
		setupFunc func()
		name      string
		call      CallConfig
		wantErr   bool
	}{
		{
			name: "valid deriv handler creation",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid http handler creation",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						URLTemplate:     "http://localhost/",
						Method:          "GET",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid processor configuration",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						URLTemplate:     "http://localhost/",
						Method:          "GET",
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid validator config",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "invalidType"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						Allow:           []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid request template",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: nil,
						Allow:           []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid url template",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:         "backend1",
						FieldsMap:    map[string]string{"field1": "value1"},
						ResponseBody: "responseBody1",
						URLTemplate:  "http://localhost/${invalid",
						Allow:        []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "circular dependency",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: map[string]any{"key1": "value1"},
						DependsOn:       []string{"backend2"},
						Allow:           []string{"allow1"},
					},
					{
						Name:            "backend2",
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: map[string]any{"key1": "value1"},
						DependsOn:       []string{"backend1"},
						Allow:           []string{"allow2"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing name in backend config",
			call: CallConfig{
				Method: "testMethod",
				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:   map[string]string{"field1": "value1"},
						URLTemplate: "http://localhost/${params.param1}",
						Allow:       []string{"allow1"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerMap := make(map[string]core.Handler)
			err := createHandler(tt.call, handlerMap)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, handlerMap, tt.call.Method)
			}
		})
	}
}
