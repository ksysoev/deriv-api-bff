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

	Client() (*clientv3.Client, error)

	Watch(ctx context.Context, cli *clientv3.Client, key string) (clientv3.WatchChan, context.CancelFunc)
}

type EtcdHandler struct {
	Conf *clientv3.Config
}

func (etcdHandler EtcdHandler) Watch(ctx context.Context, cli *clientv3.Client, key string) (clientv3.WatchChan, context.CancelFunc) {
	defer cli.Close()

	watchCtx, cancel := context.WithCancel(ctx)
	watchRespChan := cli.Watcher.Watch(watchCtx, key)

	return watchRespChan, cancel
}

func (etcdHandler EtcdHandler) Client() (*clientv3.Client, error) {
	return clientv3.New(*etcdHandler.Conf)
}

func (etcdHandler EtcdHandler) Put(ctx context.Context, cli *clientv3.Client, key, value string) error {
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, etcdHandler.Conf.DialTimeout)
	res, err := cli.Put(ctx, key, value)

	cancel()

	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("Etcd response for push config: %v", res.Header))

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
