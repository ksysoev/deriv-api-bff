package cmd

import (
	"github.com/spf13/cobra"
)

type args struct {
	appName    string
	version    string
	logLevel   string
	textFormat bool
	configPath string
}

// InitCommands initializes and returns the root command for the Backend for Frontend (BFF) service.
// It sets up the command structure and adds subcommands, including setting up persistent flags.
// It returns a pointer to a cobra.Command which represents the root command.
func InitCommands(name, version string) *cobra.Command {
	args := &args{
		appName: name,
		version: version,
	}

	cmd := &cobra.Command{
		Use:   "bff",
		Short: "Backend for Frontend service",
		Long:  "Backend for Frontend service for Deriv API",
	}

	cmd.AddCommand(ServerCommand(args))

	cmd.PersistentFlags().StringVar(&args.configPath, "config", "./runtime/config.yaml", "config file path")
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVar(&args.textFormat, "log-text", false, "log in text format, otherwise JSON")

	return cmd
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

			cfg, err := initConfig(arg.configPath)
			if err != nil {
				return err
			}

			return runServer(cmd.Context(), cfg)
		},
	}
}
