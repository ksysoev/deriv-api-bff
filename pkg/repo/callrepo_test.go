package repo

import (
	"fmt"
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
						Params: validator.Config{"param1": {Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: "template1",
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
						Params: validator.Config{"param1": {Type: "value1"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: "{{.InvalidTemplate",
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
						Params: validator.Config{"param1": {Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: "template1",
								DependsOn:       []string{"responseBody2"},
								Allow:           []string{"allow1"},
							},
							{
								FieldsMap:       map[string]string{"field2": "value2"},
								ResponseBody:    "responseBody2",
								RequestTemplate: "template2",
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
						Params: validator.Config{"param1": {Type: "string"}},
						Backend: []*BackendConfig{
							{
								FieldsMap:       map[string]string{"field1": "value1"},
								ResponseBody:    "responseBody1",
								RequestTemplate: "template1",
								DependsOn:       []string{"responseBody2"},
								Allow:           []string{"allow1"},
							},
							{
								FieldsMap:       map[string]string{"field2": "value2"},
								ResponseBody:    "responseBody2",
								RequestTemplate: "template2",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				{ResponseBody: "response1"},
				{ResponseBody: "response2"},
			},
			want: []*BackendConfig{
				{ResponseBody: "response1"},
				{ResponseBody: "response2"},
			},
			wantErr: false,
		},
		{
			name: "simple dependency",
			input: []*BackendConfig{
				{ResponseBody: "response1", DependsOn: []string{"response2"}},
				{ResponseBody: "response2"},
			},
			want: []*BackendConfig{
				{ResponseBody: "response2"},
				{ResponseBody: "response1", DependsOn: []string{"response2"}},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			input: []*BackendConfig{
				{ResponseBody: "response1", DependsOn: []string{"response2"}},
				{ResponseBody: "response2", DependsOn: []string{"response1"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "complex dependency",
			input: []*BackendConfig{
				{ResponseBody: "response1", DependsOn: []string{"response3"}},
				{ResponseBody: "response2", DependsOn: []string{"response1"}},
				{ResponseBody: "response3"},
			},
			want: []*BackendConfig{
				{ResponseBody: "response3"},
				{ResponseBody: "response1", DependsOn: []string{"response3"}},
				{ResponseBody: "response2", DependsOn: []string{"response1"}},
			},
			wantErr: false,
		},
		{
			name: "complex cicular dependency",
			input: []*BackendConfig{
				{ResponseBody: "response1", DependsOn: []string{"response3"}},
				{ResponseBody: "response2", DependsOn: []string{"response1"}},
				{ResponseBody: "response3", DependsOn: []string{"response2"}},
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
						DependsOn:       []string{"responseBody2"},
						Allow:           []string{"allow1"},
					},
					{
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: "template2",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
						DependsOn:       []string{"responseBody2"},
						Allow:           []string{"allow1"},
					},
					{
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: "template2",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "value1"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "{{.InvalidTemplate",
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

func TestUpdateCalls_ThreadSafety(t *testing.T) {
	writeDone := make(chan bool, 1)
	readDone := make(chan bool, 1)
	oldCallsConfig := &CallsConfig{
		Calls: []CallConfig{
			{
				Method: "testMethod",
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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

	newCallsConfig := func(i int) *CallsConfig {
		return &CallsConfig{
			Calls: []CallConfig{
				{
					Method: "testMethodNew",
					Params: validator.Config{"param1": {Type: "string"}},
					Backend: []*BackendConfig{
						{
							FieldsMap:       map[string]string{"field1": "value1"},
							ResponseBody:    fmt.Sprintf("responseBody0%d", i),
							RequestTemplate: "template1",
							Allow:           []string{"allow1"},
						},
					},
				},
			},
		}
	}

	go func() {
		for i := 1; i < 100; i++ {
			callsRepo.UpdateCalls(newCallsConfig(i))

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			handler := callsRepo.GetCall("testMethodNew")

			assert.NotEqualValues(t, oldHandler, handler, "old handler and new handler must be different")
		}
		writeDone <- true
	}()

	go func() {
		for i := 1; i < 100; i++ {
			handler := callsRepo.GetCall("testMethodNew")
			originalHandler := callsRepo.GetCall("testMethod")

			assert.False(t, originalHandler == nil && handler == nil, "both old and new call config cannot be missing at same time")
			assert.False(t, originalHandler != nil && handler != nil, "both old and new call config cannot exist at same time")
			assert.NotEqualValues(t, oldHandler, handler, "old handler and new handler must be different")
		}
		readDone <- true
	}()

	assert.True(t, <-writeDone)
	assert.True(t, <-readDone)
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						URLTemplate:     "http://localhost/",
						Method:          "GET",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						URLTemplate:     "http://localhost/",
						Method:          "GET",
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "invalidType"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "{{.InvalidTemplate",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:         "backend1",
						FieldsMap:    map[string]string{"field1": "value1"},
						ResponseBody: "responseBody1",
						URLTemplate:  "{{.InvalidTemplate",
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
				Params: validator.Config{"param1": {Type: "string"}},
				Backend: []*BackendConfig{
					{
						Name:            "backend1",
						FieldsMap:       map[string]string{"field1": "value1"},
						ResponseBody:    "responseBody1",
						RequestTemplate: "template1",
						DependsOn:       []string{"responseBody2"},
						Allow:           []string{"allow1"},
					},
					{
						Name:            "backend2",
						FieldsMap:       map[string]string{"field2": "value2"},
						ResponseBody:    "responseBody2",
						RequestTemplate: "template2",
						DependsOn:       []string{"responseBody1"},
						Allow:           []string{"allow2"},
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
