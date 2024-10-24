package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitCommands(t *testing.T) {
	build := "test-build"
	version := "test-version"

	cmd, err := InitCommands(build, version)
	assert.NoError(t, err)

	assert.NotNil(t, cmd)
	assert.Equal(t, "bff", cmd.Use)
	assert.Equal(t, "Backend for Frontend service", cmd.Short)
	assert.Equal(t, "Backend for Frontend service for Deriv API", cmd.Long)

	configFlag := cmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "./runtime/config.yaml", configFlag.DefValue)

	logLevelFlag := cmd.PersistentFlags().Lookup("loglevel")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	logTextFlag := cmd.PersistentFlags().Lookup("logtext")
	assert.NotNil(t, logTextFlag)
	assert.Equal(t, "false", logTextFlag.DefValue)

	subCommands := cmd.Commands()
	assert.Equal(t, 2, len(subCommands))
	assert.ElementsMatchf(t, []string{"server", "config"}, mapToNames(subCommands), "commands should match")

	configCommands := findByName(subCommands, "config").Commands()
	assert.Equal(t, 1, len(configCommands))
	assert.Equal(t, "upload", configCommands[0].Use)
}

func mapToNames(commands []*cobra.Command) []string {
	result := make([]string, len(commands))

	for i, v := range commands {
		result[i] = v.Use
	}

	return result
}

func findByName(commands []*cobra.Command, name string) *cobra.Command {
	for _, cmd := range commands {
		if cmd.Use == name {
			return cmd
		}
	}

	return nil
}

func TestServerCommand(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	arg := &args{
		build:      "test-build",
		version:    "test-version",
		ConfigPath: configPath,
		LogLevel:   "debug",
		TextFormat: true,
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

func TestConfigCommand(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	arg := &args{
		build:      "test-build",
		version:    "test-version",
		ConfigPath: configPath,
		LogLevel:   "debug",
		TextFormat: true,
	}

	cmd := ConfigCommand(arg)

	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, "Config related commands for Deriv API BFF", cmd.Short)
	assert.Equal(t, "Use this command to invoke various config related operations. Use --help for help", cmd.Long)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
}

func TestReadConfigCommand(t *testing.T) {
	configPath := createTempConfigFile(t, validConfig)

	arg := &args{
		build:      "test-build",
		version:    "test-version",
		ConfigPath: configPath,
		LogLevel:   "debug",
		TextFormat: true,
	}

	cmd := ReadConfigCommand(arg)

	assert.NotNil(t, cmd)
	assert.Equal(t, "upload", cmd.Use)
	assert.Equal(t, "Read config and push call config to etcd", cmd.Short)
	assert.Equal(t, "Read config and push call config to etcd for hot reloads. Also sets up a watcher for the config", cmd.Long)

	err := cmd.ExecuteContext(cmd.Context())

	if err.Error() != "context deadline exceeded" {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestInitCommands_BindPFlags(t *testing.T) {
	build := "test-build"
	version := "test-version"

	cmd, err := InitCommands(build, version)
	assert.NoError(t, err)

	assert.NotNil(t, cmd)

	// Check if flags are bound correctly
	assert.Equal(t, "info", viper.GetString("LOGLEVEL"))
	assert.Equal(t, false, viper.GetBool("LOGTEXT"))
	assert.Equal(t, "./runtime/config.yaml", viper.GetString("CONFIG"))
}

func TestInitCommands_UnmarshalArgs(t *testing.T) {
	build := "test-build"
	version := "test-version"

	cmd, err := InitCommands(build, version)
	assert.NoError(t, err)

	assert.NotNil(t, cmd)

	args := &args{}
	err = viper.Unmarshal(args)
	assert.NoError(t, err)

	assert.Equal(t, "info", args.LogLevel)
	assert.Equal(t, false, args.TextFormat)
	assert.Equal(t, "./runtime/config.yaml", args.ConfigPath)
}

func TestInitCommands_InvalidLogText(t *testing.T) {
	build := "test-build"
	version := "test-version"

	os.Setenv("LOGTEXT", "invalid")

	_, err := InitCommands(build, version)
	assert.Error(t, err)
}
