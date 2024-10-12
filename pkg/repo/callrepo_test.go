package repo

// import (
// 	"html/template"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestNewCallsRepository(t *testing.T) {
// 	tests := []struct {
// 		cfg     *CallsConfig
// 		name    string
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid config",
// 			cfg: &CallsConfig{
// 				Calls: []CallConfig{
// 					{
// 						Method: "testMethod",
// 						Params: map[string]string{"param1": "value1"},
// 						Backend: []BackendConfig{
// 							{
// 								FieldsMap:       map[string]string{"field1": "value1"},
// 								ResponseBody:    "responseBody1",
// 								RequestTemplate: "template1",
// 								Allow:           []string{"allow1"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "invalid template",
// 			cfg: &CallsConfig{
// 				Calls: []CallConfig{
// 					{
// 						Method: "testMethod",
// 						Params: map[string]string{"param1": "value1"},
// 						Backend: []BackendConfig{
// 							{
// 								FieldsMap:       map[string]string{"field1": "value1"},
// 								ResponseBody:    "responseBody1",
// 								RequestTemplate: "{{.InvalidTemplate",
// 								Allow:           []string{"allow1"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := NewCallsRepository(tt.cfg)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, got)
// 				assert.Contains(t, got.calls, "testMethod")

// 				callConfig := got.calls["testMethod"]
// 				assert.NotNil(t, callConfig)
// 				assert.Contains(t, callConfig.Requests, "responseBody1")

// 				requestConfig := callConfig.Requests["responseBody1"]
// 				assert.NotNil(t, requestConfig)
// 				assert.Equal(t, "responseBody1", requestConfig.ResponseBody)
// 				assert.Equal(t, []string{"allow1"}, requestConfig.Allow)
// 				assert.Equal(t, map[string]string{"field1": "value1"}, requestConfig.FieldMap)

// 				tmpl, err := template.New("request").Parse("template1")
// 				assert.NoError(t, err)
// 				assert.Equal(t, tmpl.Tree.Root.String(), requestConfig.Tmplt.Tree.Root.String())
// 			}
// 		})
// 	}
// }
// func TestGetCall(t *testing.T) {
// 	repo, err := NewCallsRepository(&CallsConfig{
// 		Calls: []CallConfig{
// 			{
// 				Method: "testMethod",
// 				Params: map[string]string{"param1": "value1"},
// 				Backend: []BackendConfig{
// 					{
// 						FieldsMap:       map[string]string{"field1": "value1"},
// 						ResponseBody:    "responseBody1",
// 						RequestTemplate: "template1",
// 						Allow:           []string{"allow1"},
// 					},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, repo)

// 	tests := []struct {
// 		name   string
// 		method string
// 		found  bool
// 	}{
// 		{
// 			name:   "existing method",
// 			method: "testMethod",
// 			found:  true,
// 		},
// 		{
// 			name:   "non-existing method",
// 			method: "nonExistingMethod",
// 			found:  false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			callConfig := repo.GetCall(tt.method)

// 			if tt.found {
// 				assert.NotNil(t, callConfig)
// 			} else {
// 				assert.Nil(t, callConfig)
// 			}
// 		})
// 	}
// }
