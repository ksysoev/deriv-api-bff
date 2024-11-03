package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/config/source"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/spf13/viper"
)

type Config struct {
	Otel      OtelConfig    `mapstructure:"otel"`
	APISource source.Config `mapstructure:"api_source"`
	Deriv     deriv.Config  `mapstructure:"deriv"`
	Server    api.Config    `mapstructure:"server"`
}

// initConfig initializes the configuration by reading from the specified config file.
// It takes configPath of type string which is the path to the configuration file.
// It returns a pointer to a config struct and an error.
// It returns an error if the configuration file cannot be read or if the configuration cannot be unmarshaled.
func initConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return &cfg, nil
}

// uploadConfig uploads the configuration using the provided context and configuration object.
// It takes ctx of type context.Context and cfg of type *Config.
// It returns an error if the configuration service creation fails, handlers loading fails, or pushing the configuration fails.
func uploadConfig(ctx context.Context, cfg *Config) error {
	calls := repo.NewCallsRepository()
	requestHandler := core.NewService(calls, nil, nil)

	sourceOpts, err := source.CreateOptions(&cfg.APISource)
	if err != nil {
		return fmt.Errorf("failed to create config source: %w", err)
	}

	svc, err := config.New(requestHandler, sourceOpts...)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := svc.LoadHandlers(ctx); err != nil {
		return fmt.Errorf("failed to load handlers: %w", err)
	}

	return svc.PutConfig(ctx)
}
