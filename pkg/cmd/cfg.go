package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/spf13/viper"
)

type Config struct {
	Otel      OtelConfig          `mapstructure:"otel"`
	Deriv     deriv.Config        `mapstructure:"deriv"`
	Server    api.Config          `mapstructure:"server"`
	APISource config.SourceConfig `mapstructure:"api_source"`
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

// putCallConfigToEtcd loads the calls config from a specific config file and pushes it to Etcd
// The function also accepts etcd settings like host and dial timeout.
// The function will return the Etcd key it had used to put the CallConfig
// The function may return empty key and an error in case of any errors.
func putConfig(ctx context.Context, cfg *Config) error {
	calls := repo.NewCallsRepository()

	requestHandler := core.NewService(calls, nil, nil)

	svc, err := config.New(cfg.APISource, requestHandler)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}

	if err := svc.LoadHandlers(ctx); err != nil {
		return fmt.Errorf("failed to load handlers: %w", err)
	}

	if err := svc.PutConfig(ctx); err != nil {
		return fmt.Errorf("failed to push config: %w", err)
	}

	return nil
}
