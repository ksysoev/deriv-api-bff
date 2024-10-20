package cmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestInitCommands(t *testing.T) {
	build := "test-build"
	version := "test-version"

	cmd := InitCommands(build, version)

	assert.NotNil(t, cmd)
	assert.Equal(t, "bff", cmd.Use)
	assert.Equal(t, "Backend for Frontend service", cmd.Short)
	assert.Equal(t, "Backend for Frontend service for Deriv API", cmd.Long)

	configFlag := cmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "./runtime/config.yaml", configFlag.DefValue)

	logLevelFlag := cmd.PersistentFlags().Lookup("log-level")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	logTextFlag := cmd.PersistentFlags().Lookup("log-text")
	assert.NotNil(t, logTextFlag)
	assert.Equal(t, "false", logTextFlag.DefValue)

	subCommands := cmd.Commands()
	assert.Equal(t, 2, len(subCommands))
	assert.ElementsMatchf(t, []string{"server", "config"}, mapToNames(subCommands), "commands should match")
}

func mapToNames(commands []*cobra.Command) []string {
	result := make([]string, len(commands))

	for i, v := range commands {
		result[i] = v.Use
	}

	return result
}

func TestServerCommand(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	arg := &args{
		build:      "test-build",
		version:    "test-version",
		configPath: configPath,
		logLevel:   "debug",
		textFormat: true,
	}

	cmd := ServerCommand(arg)

	assert.NotNil(t, cmd)
	assert.Equal(t, "server", cmd.Use)
	assert.Equal(t, "Start BFF server", cmd.Short)
	assert.Equal(t, "Start BFF server for Deriv API", cmd.Long)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
}
