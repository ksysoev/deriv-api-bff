package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
	"github.com/spf13/viper"
)

type config struct {
	API handlers.Config `mapstructure:"api"`
}

func initConfig(configPath string) (*config, error) {
	viper.SetConfigFile(configPath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	encdedCfg, _ := json.MarshalIndent(cfg, "", "  ")
	slog.Debug("config:\n" + string(encdedCfg))

	return cfg, nil
}
