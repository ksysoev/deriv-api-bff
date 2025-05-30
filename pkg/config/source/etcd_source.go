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

const (
	defaultTimeoutSeconds = 5
	defaultReducerTimeout = 1 * time.Second
)

type EtcdConfig struct {
	Prefix  string `mapstructure:"prefix"`
	Servers string `mapstructure:"servers"`
}

type EtcdSource struct {
	cli             *clientv3.Client
	prefix          string
	reducerInterval time.Duration
}

// NewEtcdSource creates a new EtcdSource instance configured with the provided EtcdConfig.
// It takes cfg of type EtcdConfig which includes the servers and prefix for the etcd client.
// It returns a pointer to EtcdSource and an error.
// It returns an error if no etcd key prefix is provided or if the etcd client creation fails.
func NewEtcdSource(cfg EtcdConfig) (*EtcdSource, error) {
	serves := strings.Split(cfg.Servers, ",")

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
		prefix:          cfg.Prefix,
		cli:             cli,
		reducerInterval: defaultReducerTimeout,
	}, nil
}

// LoadConfig loads configuration data from an etcd source.
// It takes a context.Context as a parameter to manage the request's lifetime.
// It returns a slice of handlerfactory.Config and an error.
// It returns an error if it fails to get the config from etcd or if it fails to unmarshal the config data.
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

// PutConfig stores the provided configuration in the etcd key-value store.
// It takes a context ctx of type context.Context and a slice of handlerfactory.Config cfg.
// It returns an error if marshalling the config fails or if putting the config into etcd fails.
// The function creates a context with a timeout for each put operation.
func (es *EtcdSource) PutConfig(ctx context.Context, cfg []handlerfactory.Config) error {
	//TODO: add logic for removing keys that are not in the new config
	indx := make(map[string]struct{}, len(cfg))

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

		indx[c.Method] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
	defer cancel()

	data, err := es.cli.Get(ctx, es.prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get config from etcd: %w", err)
	}

	for _, kv := range data.Kvs {
		key := strings.TrimPrefix(string(kv.Key), es.prefix)
		if _, ok := indx[key]; !ok {
			ctx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
			_, err := es.cli.Delete(ctx, string(kv.Key))

			cancel()

			if err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}
		}
	}

	return nil
}

// Watch monitors changes to keys with a specified prefix in an etcd cluster and triggers an update callback.
// It takes a context.Context and a callback function onUpdate which is called when changes are detected.
// It does not return any values.
// The function continues to watch for changes until the context is canceled.
func (es *EtcdSource) Watch(ctx context.Context, onUpdate func()) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	reducerOnUpdate := makeReducer(ctx, onUpdate, es.reducerInterval)

	rch := es.cli.Watch(ctx, es.prefix, clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			return
		case wresp := <-rch:
			for range wresp.Events {
				reducerOnUpdate()
			}
		}
	}
}

// makeReducer creates a function that triggers an update at a specified interval.
// It takes a context 'ctx' of type context.Context, an 'onUpdate' function to be called on update, and an 'interval' of type time.Duration.
// It returns a function that can be called to signal an update.
func makeReducer(ctx context.Context, onUpdate func(), interval time.Duration) func() {
	updates := make(chan struct{}, 1)

	go func() {
		timer := time.NewTimer(interval)
		defer timer.Stop()

		var timerChan <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-updates:
				timer.Reset(interval)
				timerChan = timer.C

			case <-timerChan:
				onUpdate()

				timerChan = nil
			}
		}
	}()

	return func() {
		select {
		case updates <- struct{}{}:
		default:
		}
	}
}
