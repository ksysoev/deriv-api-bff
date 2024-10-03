package cmd

import (
	"fmt"
	"strings"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/spf13/viper"
)

type config struct {
	Server api.Config      `mapstructure:"server"`
	API    handlers.Config `mapstructure:"api"`
	Deriv  deriv.Config    `mapstructure:"deriv"`
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

	return cfg, nil
}
