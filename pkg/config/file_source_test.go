package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var validConfig = `
server:
  listen: ":0"
deriv:
  endpoint: "wss://localhost/"
api:
  calls: []
etcd:
  dialTimeoutSeconds: 1
  servers: ["host1:port1", "host2:port2"]
`

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	tmpdir := os.TempDir()
	configPath := tmpdir + "/" + t.Name() + "_test_config.yaml"

	err := os.WriteFile(configPath, []byte(content), 0o600)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(configPath)
	})

	return configPath
}

func TestFileSource_Init(t *testing.T) {
	// Setup
	viper.Reset()

	fileSource := NewFileSource(createTempConfigFile(t, validConfig))

	// Test Init
	err := fileSource.Init()
	assert.NoError(t, err)

	// Test GetConfigurations
	config, err := fileSource.GetConfigurations()

	assert.NoError(t, err)
	assert.NotNil(t, config)
}

func TestFileSource_WatchConfig_Pass(t *testing.T) {
	// Setup
	fileSource := NewFileSource(createTempConfigFile(t, validConfig))
	event := NewEvent[any]()

	event.RegisterHandler(func(_ context.Context, _ any) {})

	// Test WatchConfig
	err := fileSource.WatchConfig(event, "key1")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(fileSource.GetWatchKeys()))
}

func TestFileSource_WatchConfig_Error(t *testing.T) {
	// Setup
	fileSource := NewFileSource(createTempConfigFile(t, validConfig))
	event := NewEvent[any]()

	// Test WatchConfig
	err := fileSource.WatchConfig(event, "key1")

	assert.Error(t, err)
	assert.Equal(t, 0, len(fileSource.GetWatchKeys()))
}

func TestFileSource_onFileChange(t *testing.T) {
	// Setup
	viper.Reset()

	configFile := createTempConfigFile(t, "")
	fileSource := NewFileSource(configFile)

	err := fileSource.Init()

	assert.NoError(t, err)

	oldConfig, err := fileSource.GetConfigurations()

	assert.NoError(t, err)
	assert.Empty(t, oldConfig.Deriv.Endpoint)
	assert.Empty(t, oldConfig.API.Calls)

	createTempConfigFile(t, validConfig)

	// need to give some time for config to refresh
	time.Sleep(1 * time.Second)

	// Check if the configuration was updated
	config, err := fileSource.GetConfigurations()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, config.API.Calls)
	assert.NotEmpty(t, config.Deriv.Endpoint)
	assert.Equal(t, "wss://localhost/", config.Deriv.Endpoint)
}

func TestFileSource_GetPriority(t *testing.T) {
	fileSource := NewFileSource("test_config.yaml")

	assert.Equal(t, P1, fileSource.GetPriority())
}

func TestFileSource_Name(t *testing.T) {
	fileSource := NewFileSource("test_config.yaml")

	assert.Equal(t, "VIPER_FILE_SOURCE", fileSource.Name())
}

func TestFileSource_Close(t *testing.T) {
	fileSource := NewFileSource("test_config.yaml")

	assert.NoError(t, fileSource.Close())
}
