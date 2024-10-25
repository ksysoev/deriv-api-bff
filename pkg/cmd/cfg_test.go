package cmd

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/stretchr/testify/assert"
)

var validConfig = `
server:
  listen: ":0"
deriv:
  endpoint: "wss://localhost/"
api:
  calls:
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

func TestInitConfig_Valid(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	cfg, err := initConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, ":0", cfg.Server.Listen)
	assert.Equal(t, "wss://localhost/", cfg.Deriv.Endpoint)
	assert.Equal(t, 1, cfg.Etcd.DialTimeoutSeconds)
	assert.ElementsMatch(t, []string{"host1:port1", "host2:port2"}, cfg.Etcd.Servers)
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

func TestPutCallConfig_Success(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)
	mockEtcd := repo.NewMockEtcd(t)

	mockEtcd.EXPECT().Put("CallConfig", "null").Return(nil)

	err := putCallConfigToEtcd(mockEtcd, configPath)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPutCallConfig_Fail_OnPut(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)
	mockEtcd := repo.NewMockEtcd(t)
	expectedErr := errors.New("test error")

	mockEtcd.EXPECT().Put("CallConfig", "null").Return(expectedErr)

	err := putCallConfigToEtcd(mockEtcd, configPath)

	if err != expectedErr {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPutCallConfig_Fail_OnConfigRead(t *testing.T) {
	configPath := createTempConfigFile(t, "invalid content")
	mockEtcd := repo.NewMockEtcd(t)
	expectedErr := "failed to read config:"

	err := putCallConfigToEtcd(mockEtcd, configPath)

	if !strings.HasPrefix(err.Error(), expectedErr) {
		t.Errorf("Unexpected error: %s", err)
	}
}
