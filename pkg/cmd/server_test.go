package cmd

var validPartialConfig = `
server:
  listen: ":0"
`

// func TestRunServer(t *testing.T) {
// 	cfg := &config.Config{
// 		Server: api.Config{
// 			Listen: ":0",
// 		},
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel()

// 	err := runServer(ctx, cfg)

// 	assert.NoError(t, err)
// }

// func TestRunServer_WithWatcher(t *testing.T) {
// 	cfg := &config.Config{}
// 	source := config.NewFileSource(createTempConfigFile(t, validPartialConfig))

// 	err := source.Init(cfg)
// 	assert.NoError(t, err)

// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel()

// 	err = runServer(ctx, cfg)

// 	assert.Nil(t, cfg.API.Calls)
// 	assert.NoError(t, err)

// 	callsConfig := config.CallsConfig{
// 		Calls: []config.CallConfig{
// 			{
// 				Method: "testMethod",
// 				Params: validator.Config{"param1": validator.FieldSchema{Type: "string"}},
// 				Backend: []*config.BackendConfig{
// 					{
// 						FieldsMap:       map[string]string{"field1": "value1"},
// 						ResponseBody:    "responseBody1",
// 						RequestTemplate: map[string]any{"template1": "t1"},
// 						Allow:           []string{"allow1"},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	newCfg := &config.Config{
// 		Server: cfg.Server,
// 		API:    callsConfig,
// 	}
// 	newCfgJSON, err := json.Marshal(newCfg)
// 	assert.NoError(t, err)

// 	newCfgJSONString := string(newCfgJSON)

// 	createTempConfigFile(t, newCfgJSONString)
// 	time.Sleep(1 * time.Second)

// 	newCfg, err = source.GetConfigurations()
// 	watchKeys := source.GetWatchKeys()

// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, watchKeys)
// 	assert.NotEmpty(t, newCfg.API.Calls)
// 	assert.NotNil(t, watchKeys["api.calls"])
// }

// func TestRunServer_Error(t *testing.T) {
// 	cfg := &config.Config{
// 		Server: api.Config{
// 			Listen: ":0",
// 		},
// 		API: config.CallsConfig{
// 			Calls: []config.CallConfig{
// 				{
// 					Method: "GET",
// 					Params: validator.Config{
// 						"param": &validator.FieldSchema{
// 							Type: "InvalidType",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel()

// 	err := runServer(ctx, cfg)

// 	assert.Error(t, err)
// }
