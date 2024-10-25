package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/spf13/viper"
)

type config struct {
	Server api.Config       `mapstructure:"server"`
	Deriv  deriv.Config     `mapstructure:"deriv"`
	API    repo.CallsConfig `mapstructure:"api"`
	Etcd   repo.EtcdConfig  `mapstructure:"etcd"`
}

// initConfig initializes the configuration by reading from the specified config file.
// It takes configPath of type string which is the path to the configuration file.
// It returns a pointer to a config struct and an error.
// It returns an error if the configuration file cannot be read or if the configuration cannot be unmarshaled.
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

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return cfg, nil
}

// putCallConfigToEtcd loads the calls config from a specific config file and pushes it to Etcd
// The function also accepts etcd settings like host and dial timeout.
// The function will return the Etcd key it had used to put the CallConfig
// The function may return empty key and an error in case of any errors.
func putCallConfigToEtcd(etcdHandler repo.Etcd, configPath string) error {
	cfg, err := initConfig(configPath)

	if err != nil {
		return err
	}

	callConfig := cfg.API.Calls

	callConfigJSON, err := json.Marshal(callConfig)

	if err != nil {
		return err
	}

	err = etcdHandler.Put("CallConfig", string(callConfigJSON))

	if err != nil {
		return err
	}

	return nil
}
