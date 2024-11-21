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
func initConfig(arg *args) (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	if arg.ConfigPath != "" {
		v.SetConfigFile(arg.ConfigPath)

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyArgsToConfig(arg, &cfg)

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

// verifyConfig verifies the provided configuration.
// It takes ctx of type context.Context and cfg of type *Config.
// It returns an error if the configuration is invalid or if there is a failure in creating the config source or service.
// It returns nil if the configuration is successfully verified and handlers are loaded.
func verifyConfig(ctx context.Context, cfg *Config) error {
	sourceOpts, err := source.CreateOptions(&cfg.APISource)
	if err != nil {
		return fmt.Errorf("failed to create config source: %w", err)
	}

	calls := repo.NewCallsRepository()
	requestHandler := core.NewService(calls, nil, nil)

	svc, err := config.New(requestHandler, sourceOpts...)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := svc.LoadHandlers(ctx); err != nil {
		return fmt.Errorf("failed to load handlers: %w", err)
	}

	slog.Info("Configuration is valid")

	return nil
}

// downloadConfig downloads the configuration from a specified source and writes it to a given path.
// It takes a context.Context, a pointer to Config, and a string representing the path to write the configuration.
// It returns an error if there is a failure in creating the config source, creating the config service, or writing the config.
func downloadConfig(ctx context.Context, cfg *Config, path string) error {
	sourceOpts, err := source.CreateOptions(&cfg.APISource)
	if err != nil {
		return fmt.Errorf("failed to create config source: %w", err)
	}

	calls := repo.NewCallsRepository()
	requestHandler := core.NewService(calls, nil, nil)

	svc, err := config.New(requestHandler, sourceOpts...)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	err = svc.WriteConfig(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// applyArgsToConfig applies command-line arguments to the configuration.
// It takes arg of type *args and cfg of type *Config.
// It does not return any values.
// If any of the fields in arg are non-empty, it updates the corresponding fields in cfg.
func applyArgsToConfig(arg *args, cfg *Config) {
	if arg.apiSourcePath != "" {
		cfg.APISource.Path = arg.apiSourcePath
	}

	if arg.apiSourceEtcdServers != "" {
		cfg.APISource.Etcd.Servers = arg.apiSourceEtcdServers
	}

	if arg.apiSourceEtcdPrefix != "" {
		cfg.APISource.Etcd.Prefix = arg.apiSourceEtcdPrefix
	}
}
