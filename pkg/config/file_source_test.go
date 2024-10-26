package config

import (
	"os"
	"testing"

	"github.com/fsnotify/fsnotify"
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
	configPath := tmpdir + "/test_config.yaml"

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

func TestFileSource_WatchConfig(t *testing.T) {
	// Setup
	fileSource := NewFileSource(createTempConfigFile(t, validConfig))

	// Test WatchConfig
	fileSource.WatchConfig("key1")
	assert.Contains(t, *fileSource.watchKeyPrefixSet, "key1")
}

func TestFileSource_onFileChange(t *testing.T) {
	// Setup
	viper.Reset()

	configFile := createTempConfigFile(t, validConfig)
	fileSource := NewFileSource(configFile)

	err := fileSource.Init()

	assert.NoError(t, err)

	fileSource.watchKeyPrefixSet = &map[string]struct{}{".API": {}}

	createTempConfigFile(t, validConfig)

	// Test onFileChange
	event := fsnotify.Event{Name: configFile}

	fileSource.onFileChange(event)

	// Check if the configuration was updated
	config, err := fileSource.GetConfigurations()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, config.API.Calls)
	assert.NotNil(t, config.Deriv.Endpoint)
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
