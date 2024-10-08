package cmd

import (
	"github.com/spf13/cobra"
)

// InitCommands initializes and returns the root command for the Backend for Frontend (BFF) service.
// It sets up the command structure and adds subcommands, including setting up persistent flags.
// It returns a pointer to a cobra.Command which represents the root command.
func InitCommands() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "bff",
		Short: "Backend for Frontend service",
		Long:  "Backend for Frontend service for Deriv API",
	}

	cmd.AddCommand(ServerCommand(&configPath))

	cmd.PersistentFlags().StringVar(&configPath, "config", "./runtime/config.yaml", "config file path")

	return cmd
}

// ServerCommand creates a new cobra.Command to start the BFF server for Deriv API.
// It takes cfgPath of type *string which is the path to the configuration file.
// It returns a pointer to a cobra.Command which can be executed to start the server.
// It returns an error if the logger initialization fails, the configuration cannot be loaded, or the server fails to run.
func ServerCommand(cfgPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start BFF server",
		Long:  "Start BFF server for Deriv API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := initLogger(); err != nil {
				return err
			}

			cfg, err := initConfig(*cfgPath)
			if err != nil {
				return err
			}

			return runServer(cmd.Context(), cfg)
		},
	}
}
