package cmd

import (
	"github.com/spf13/cobra"
)

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
