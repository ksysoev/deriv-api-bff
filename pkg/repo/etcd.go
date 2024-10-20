package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Etcd interface {
	Put(ctx context.Context, cli *clientv3.Client, key string, value string) error
	Client(cfg EtcdConfig) (*clientv3.Client, error)
}

type EtcdHandler struct {
	Conf *clientv3.Config
}

func (etcdHandler EtcdHandler) Client(cfg EtcdConfig) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   cfg.Servers,
		DialTimeout: time.Duration(cfg.DialTimeoutSeconds * int(time.Second)),
	})
}

func (etcdHandler EtcdHandler) Put(ctx context.Context, cli *clientv3.Client, key, value string) error {
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, etcdHandler.Conf.DialTimeout)
	res, err := cli.Put(ctx, key, value)

	slog.Debug(fmt.Sprintf("Etcd response for push config: %v", res.Header))

	cancel()

	if err != nil {
		return err
	}

	return nil
}

func NewEtcdHandler(etcdConfig EtcdConfig) Etcd {
	return &EtcdHandler{
		Conf: &clientv3.Config{
			Endpoints:   etcdConfig.Servers,
			DialTimeout: time.Duration(etcdConfig.DialTimeoutSeconds * int(time.Second)),
		},
	}
}
