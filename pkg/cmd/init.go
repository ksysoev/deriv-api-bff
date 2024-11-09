package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type args struct {
	build                string
	version              string
	LogLevel             string `mapstructure:"loglevel"`
	ConfigPath           string `mapstructure:"config"`
	apiSourcePath        string
	apiSourceEtcdServers string
	apiSourceEtcdPrefix  string
	TextFormat           bool `mapstructure:"logtext"`
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

	configCmd := ConfigCommand(args)
	configCmd.AddCommand(UploadConfigCommand(args))
	configCmd.AddCommand(VerifyConfigCommand(args))
	cmd.AddCommand(configCmd)

	cmd.PersistentFlags().StringVar(&args.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&args.LogLevel, "loglevel", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&args.TextFormat, "logtext", false, "log in text format, otherwise JSON")
	cmd.PersistentFlags().StringVar(&args.apiSourcePath, "api-source-path", "", "path to the API source file")
	cmd.PersistentFlags().StringVar(&args.apiSourceEtcdServers, "api-source-etcd-servers", "", "etcd servers for API source")
	cmd.PersistentFlags().StringVar(&args.apiSourceEtcdPrefix, "api-source-etcd-prefix", "", "etcd prefix for API source")

	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, fmt.Errorf("failed to parse env args: %w", err)
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

			cfg, err := initConfig(arg)
			if err != nil {
				return err
			}

			if err := initMetricProvider(cmd.Context(), &cfg.Otel); err != nil {
				return err
			}

			return runServer(cmd.Context(), cfg)
		},
	}
}

// ConfigCommand is a top-level cobra.Command for config relation operations for Deriv API.
// You can use its sub commands as `config [sub command]`
func ConfigCommand(_ *args) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Config related commands for Deriv API BFF",
		Long:  "Use this command to invoke various config related operations. Use --help for help",
		RunE: func(_ *cobra.Command, _ []string) error {
			return fmt.Errorf("no subcommand provided")
		},
	}
}

// UploadConfigCommand creates a new cobra.Command to load the calls config for Deriv API.
// The config is loaded and then pushed to etcd for watching changes.
// It can take cfgPath of type *string which is the path to the configuration file as an argument.
// It also takes the etcd host URL and dial timeout in seconds as argument
// It returns a pointer to a cobra.Command which can be executed to load the config.
// It returns an error if the logger initialization fails, the configuration cannot be loaded, or there is error thrown by etcd.
func UploadConfigCommand(arg *args) *cobra.Command {
	return &cobra.Command{
		Use:   "upload",
		Short: "Read config and push call config to etcd",
		Long:  "Read config and push call config to etcd for hot reloads. Also sets up a watcher for the config",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := initLogger(arg); err != nil {
				return err
			}

			slog.Info("Trying to load config...", slog.String("version", arg.version), slog.String("build", arg.build))

			cfg, err := initConfig(arg)

			if err != nil {
				return err
			}

			return uploadConfig(cmd.Context(), cfg)
		},
	}
}

// VerifyConfigCommand creates a new cobra.Command to verify the configuration for correctness.
// It takes arg of type *args which contains the necessary parameters for the command.
// It returns a pointer to a cobra.Command which can be executed to perform the verification.
// It returns an error if initializing the logger or configuration fails, or if the configuration verification fails.
func VerifyConfigCommand(arg *args) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify the config",
		Long:  "Verify the config for correctness",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := initLogger(arg); err != nil {
				return err
			}

			slog.Info("Verifying config...", slog.String("version", arg.version), slog.String("build", arg.build))

			cfg, err := initConfig(arg)

			if err != nil {
				return err
			}

			return verifyConfig(cmd.Context(), cfg)
		},
	}
}
