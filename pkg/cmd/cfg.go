package cmd

import (
	"encoding/json"

	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
)

// initConfig initializes the configuration by reading from the specified config file.
// It takes configPath of type string which is the path to the configuration file.
// It returns a pointer to a config struct and an error.
// It returns an error if the configuration file cannot be read or if the configuration cannot be unmarshaled.
func initConfig(configPath string) (*config.Config, error) {
	fileSource := config.NewFileSource(configPath)

	err := fileSource.Init()
	if err != nil {
		return nil, err
	}

	return fileSource.GetConfigurations()
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
