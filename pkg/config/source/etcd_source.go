package source

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
)

const defaultTimeoutSeconds = 5

type EtcdConfig struct {
	Prefix  string `mapstructure:"prefix"`
	Servers string `mapstructure:"servers"`
}

type EtcdSource struct {
	cli    *clientv3.Client
	prefix string
}

func NewEtcdSource(cfg EtcdConfig) (*EtcdSource, error) {
	serves := strings.Split(cfg.Servers, ",")

	if len(serves) == 0 {
		return nil, fmt.Errorf("no etcd servers provided")
	}

	if cfg.Prefix == "" {
		return nil, fmt.Errorf("no etcd key prefix provided")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   serves,
		DialTimeout: defaultTimeoutSeconds * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdSource{
		prefix: cfg.Prefix,
		cli:    cli,
	}, nil
}

func (es *EtcdSource) LoadConfig(ctx context.Context) ([]handlerfactory.Config, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
	defer cancel()

	data, err := es.cli.Get(ctx, es.prefix, clientv3.WithPrefix())

	if err != nil {
		return nil, fmt.Errorf("failed to get config from etcd: %w", err)
	}

	cfg := make([]handlerfactory.Config, 0, data.Count)

	for _, kv := range data.Kvs {
		var c handlerfactory.Config
		err := json.Unmarshal(kv.Value, &c)

		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		cfg = append(cfg, c)
	}

	return cfg, nil
}

func (es *EtcdSource) PutConfig(ctx context.Context, cfg []handlerfactory.Config) error {
	//TODO: add logic for removing keys that are not in the new config
	for _, c := range cfg {
		data, err := json.Marshal(c)

		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
		_, err = es.cli.Put(ctx, es.prefix+c.Method, string(data))

		cancel()

		if err != nil {
			return fmt.Errorf("failed to put config: %w", err)
		}
	}

	return nil
}
