package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Etcd interface {
	Put(ctx context.Context, key string, value string) error

	// TODO: implement watcher
	Watch(ctx context.Context, key string)
}

type EtcdHandler struct {
	conf *clientv3.Config
}

func NewEtcdHandler(etcdURL string, dialTimeout time.Duration) *EtcdHandler {
	return &EtcdHandler{
		conf: &clientv3.Config{
			Endpoints:   []string{etcdURL},
			DialTimeout: dialTimeout,
		},
	}
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
