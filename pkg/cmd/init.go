package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type args struct {
	build      string
	version    string
	LogLevel   string `mapstructure:"LOG_LEVEL"`
	ConfigPath string `mapstructure:"CONFIG"`
	TextFormat bool   `mapstructure:"LOG_TEXT"`
}

// InitCommands initializes and returns the root command for the Backend for Frontend (BFF) service.
// It sets up the command structure and adds subcommands, including setting up persistent flags.
// It returns a pointer to a cobra.Command which represents the root command.
func InitCommands(build, version string) (*cobra.Command, error) {
	args := &args{
		build:   build,
		version: version,
	}

	cmd := &cobra.Command{
		Use:   "bff",
		Short: "Backend for Frontend service",
		Long:  "Backend for Frontend service for Deriv API",
	}

	cmd.AddCommand(ServerCommand(args))

	cmd.PersistentFlags().StringVar(&args.ConfigPath, "config", "./runtime/config.yaml", "config file path")
	cmd.PersistentFlags().StringVar(&args.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&args.TextFormat, "log-text", false, "log in text format, otherwise JSON")

	if err := viper.BindPFlag("LOG_LEVEL", cmd.PersistentFlags().Lookup("log-level")); err != nil {
		return nil, fmt.Errorf("failed to bind log level flag: %w", err)
	}

	if err := viper.BindPFlag("LOG_TEXT", cmd.PersistentFlags().Lookup("log-text")); err != nil {
		return nil, fmt.Errorf("failed to bind log text flag: %w", err)
	}

	if err := viper.BindPFlag("CONFIG", cmd.PersistentFlags().Lookup("config")); err != nil {
		return nil, fmt.Errorf("failed to bind config flag: %w", err)
	}

	viper.AutomaticEnv()

	if err := viper.Unmarshal(args); err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}

	return cmd, nil
}

// ServerCommand creates a new cobra.Command to start the BFF server for Deriv API.
// It takes cfgPath of type *string which is the path to the configuration file.
// It returns a pointer to a cobra.Command which can be executed to start the server.
// It returns an error if the logger initialization fails, the configuration cannot be loaded, or the server fails to run.
func ServerCommand(arg *args) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start BFF server",
		Long:  "Start BFF server for Deriv API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := initLogger(arg); err != nil {
				return err
			}

			slog.Info("Starting Deriv API BFF server", slog.String("version", arg.version), slog.String("build", arg.build))

			cfg, err := initConfig(arg.ConfigPath)
			if err != nil {
				return err
			}

			return runServer(cmd.Context(), cfg)
		},
	}
}
