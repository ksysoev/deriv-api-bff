package source

import "github.com/ksysoev/deriv-api-bff/pkg/config"

type Config struct {
	Etcd EtcdConfig `mapstructure:"etcd"`
	Path string     `mapstructure:"path"`
}

// CreateOptions generates a slice of configuration options based on the provided Config.
// It takes cfg of type *Config.
// It returns a slice of config.Option and an error.
// It returns an error if creating the Etcd source fails.
func CreateOptions(cfg *Config) ([]config.Option, error) {
	opts := make([]config.Option, 0, 2)

	if cfg.Path != "" {
		fs := NewFileSource(cfg.Path)
		opts = append(opts, config.WithLocalSource(fs))
	}

	if cfg.Etcd.Servers != "" {
		es, err := NewEtcdSource(cfg.Etcd)
		if err != nil {
			return nil, err
		}

		opts = append(opts, config.WithRemoteSource(es))
	}

	return opts, nil
}
