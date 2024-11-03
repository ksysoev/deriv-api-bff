package source

import "github.com/ksysoev/deriv-api-bff/pkg/config"

type Config struct {
	Etcd EtcdConfig `mapstructure:"etcd"`
	Path string     `mapstructure:"path"`
}

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
