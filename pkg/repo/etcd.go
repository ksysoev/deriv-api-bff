package repo

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Etcd interface {
	Put(key string, value string) error

	Watch(key string) (clientv3.WatchChan, context.CancelFunc)

	Close() error
}

type EtcdHandler struct {
	cli  *clientv3.Client
	mu   *sync.RWMutex
	ctx  context.Context
	conf clientv3.Config
}

func (etcdHandler *EtcdHandler) Watch(key string) (clientv3.WatchChan, context.CancelFunc) {
	watchCtx, cancel := context.WithCancel(etcdHandler.ctx)
	watchRespChan := etcdHandler.cli.Watcher.Watch(watchCtx, key)

	return watchRespChan, cancel
}

func (etcdHandler *EtcdHandler) Put(key, value string) error {
	etcdHandler.mu.Lock()
	defer etcdHandler.mu.Unlock()

	ctx, cancel := context.WithTimeout(etcdHandler.ctx, etcdHandler.conf.DialTimeout)
	res, err := etcdHandler.cli.Put(ctx, key, value)

	cancel()

	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("Etcd response for push config: %v", res.Header))

	return nil
}

func (etcdHandler *EtcdHandler) Close() error {
	return etcdHandler.cli.Close()
}

func NewEtcdHandler(ctx context.Context, etcdConfig EtcdConfig) (Etcd, error) {
	conf := clientv3.Config{
		Endpoints:   etcdConfig.Servers,
		DialTimeout: time.Duration(etcdConfig.DialTimeoutSeconds * int(time.Second)),
	}
	cli, err := clientv3.New(conf)

	if err != nil {
		return nil, err
	}

	return &EtcdHandler{
		conf: conf,
		cli:  cli,
		ctx:  ctx,
		mu:   &sync.RWMutex{},
	}, nil
}
