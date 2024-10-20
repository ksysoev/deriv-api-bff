package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Etcd interface {
	Put(ctx context.Context, key string, value string) error
}

type EtcdHandler struct {
	conf *clientv3.Config
}

func (etcdHandler *EtcdHandler) Put(ctx context.Context, key, value string) error {
	cli, err := clientv3.New(*etcdHandler.conf)

	if err != nil {
		return err
	}

	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, etcdHandler.conf.DialTimeout)
	res, err := cli.Put(ctx, key, value)

	slog.Debug(fmt.Sprintf("Etcd response for push config: %v", res.Header))

	cancel()

	if err != nil {
		return err
	}

	return nil
}

func NewEtcdHandler(etcdConfig EtcdConfig) *Etcd {
	return &EtcdHandler{
		conf: &clientv3.Config{
			Endpoints:   etcdConfig.Servers,
			DialTimeout: time.Duration(etcdConfig.DialTimeoutSeconds * int(time.Second)),
		},
	}
}
