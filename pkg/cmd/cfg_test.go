package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/stretchr/testify/assert"
)

var callsConfig = `
- method: "testMethod"
  backend:
    - response_body: "ping"
      allow: 
        - value
      request_template:
        ping: 1
`

var validConfig = `
server:
  listen: ":0"
deriv:
  endpoint: "wss://localhost/"
api:
  calls:
api_source:
  etcd:
    servers: "localhost:2379"
    prefix: "api::"
  path: "./runtime/calls.yaml"
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

func TestInitConfig_Valid(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	cfg, err := initConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, ":0", cfg.Server.Listen)
	assert.Equal(t, "wss://localhost/", cfg.Deriv.Endpoint)
	assert.Equal(t, "localhost:2379", cfg.APISource.Etcd.Servers)
}

func TestInitConfig_InvalidContent(t *testing.T) {
	configPath := createTempConfigFile(t, "invalid content")

	cfg, err := initConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestInitConfig_Missing(t *testing.T) {
	dir := os.TempDir()
	configPath := dir + "/missing_config.yaml"

	cfg, err := initConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestUploadConfig_Success(t *testing.T) {
	ctx := context.Background()

	path := createTempConfigFile(t, callsConfig)

	cfg := &Config{
		APISource: config.SourceConfig{
			Path: path,
		},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Equal(t, "local and remote sources are required", err.Error())
}

func TestUploadConfig_FailCreateService(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		APISource: config.SourceConfig{},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Equal(t, "failed to create config service: no configuration source provided", err.Error())
}

func TestUploadConfig_FailLoadHandlers(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		APISource: config.SourceConfig{
			Path: "invalid_path",
		},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Equal(t, "failed to load handlers: failed to load config: failed to stat file: stat invalid_path: no such file or directory", err.Error())
}
