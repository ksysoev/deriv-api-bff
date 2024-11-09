package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/config/source"
	"github.com/stretchr/testify/assert"
)

var callsConfig = `
- method: "testMethod"
  backend:
    - allow: 
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

	cfg, err := initConfig(&args{ConfigPath: configPath})
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, ":0", cfg.Server.Listen)
	assert.Equal(t, "wss://localhost/", cfg.Deriv.Endpoint)
	assert.Equal(t, "localhost:2379", cfg.APISource.Etcd.Servers)
}

func TestInitConfig_InvalidContent(t *testing.T) {
	configPath := createTempConfigFile(t, "invalid content")

	cfg, err := initConfig(&args{ConfigPath: configPath})
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestInitConfig_Missing(t *testing.T) {
	dir := os.TempDir()
	configPath := dir + "/missing_config.yaml"

	cfg, err := initConfig(&args{ConfigPath: configPath})
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestUploadConfig(t *testing.T) {
	ctx := context.Background()

	path := createTempConfigFile(t, callsConfig)

	cfg := &Config{
		APISource: source.Config{
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
		APISource: source.Config{},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Equal(t, "failed to create config service: local or remote source is required", err.Error())
}

func TestUploadConfig_FailLoadHandlers(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		APISource: source.Config{
			Path: "invalid_path",
		},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Equal(t, "failed to load handlers: failed to load config: failed to stat file: stat invalid_path: no such file or directory", err.Error())
}

func TestUploadConfig_FailCreateSource(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Etcd: source.EtcdConfig{
				Servers: "localhost:2379",
				Prefix:  "",
			},
		},
	}

	err := uploadConfig(ctx, cfg)
	assert.Error(t, err)
}
func TestVerifyConfig_Valid(t *testing.T) {
	ctx := context.Background()

	path := createTempConfigFile(t, callsConfig)

	cfg := &Config{
		APISource: source.Config{
			Path: path,
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.NoError(t, err)
}

func TestVerifyConfig_MissingSources(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Path: "",
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.Error(t, err)
}

func TestVerifyConfig_FailCreateSource(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Etcd: source.EtcdConfig{
				Servers: "localhost:2379",
				Prefix:  "",
			},
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.Error(t, err)
}

func TestVerifyConfig_FailCreateService(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Path: "invalid_path",
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestVerifyConfig_FailToCreateSources(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Path: "invalid_path",
			Etcd: source.EtcdConfig{
				Servers: "localhost:2379",
				Prefix:  "",
			},
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.Error(t, err)
}

func TestVerifyConfig_FailLoadHandlers(t *testing.T) {
	ctx := context.Background()

	cfg := &Config{
		APISource: source.Config{
			Path: "invalid_path",
		},
	}

	err := verifyConfig(ctx, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load handlers")
}
