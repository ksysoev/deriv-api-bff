package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/etcd/clientv3"
)

func TestNew_Success(t *testing.T) {
	cfg := config.EtcdConfig{
		Servers:            []string{"localhost:7000"},
		DialTimeoutSeconds: 1,
	}
	ctx := context.Background()
	etcd, err := NewEtcdHandler(ctx, cfg)

	if err != nil {
		t.Errorf("Unexpected err: %v", err)
	}

	assert.NotEmpty(t, etcd)
	assert.Implements(t, (*Etcd)(nil), etcd)
}

func TestNew_Error(t *testing.T) {
	cfg := config.EtcdConfig{}
	ctx := context.Background()
	etcd, err := NewEtcdHandler(ctx, cfg)

	if err.Error() != "etcdclient: no available endpoints" {
		t.Errorf("Unexpected err: %v", err)
	}

	assert.Empty(t, etcd)
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)
	etcd := newEtcdHandlerWithCli(ctx, cli)

	mockLease := NewMockLease(t)
	mockWatcher := NewMockWatcher(t)

	cli.Lease = mockLease
	cli.Watcher = mockWatcher

	mockWatcher.EXPECT().Close().Return(nil)
	mockLease.EXPECT().Close().Return(nil)

	err := etcd.Close()

	if err.Error() != "context canceled" {
		t.Errorf("Unexpected error: %v", err)
	}

	assert.Equal(t, 1, len(mockWatcher.Calls))
	assert.Equal(t, 1, len(mockLease.Calls))
}

func TestPut_Success(t *testing.T) {
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)
	etcd := newEtcdHandlerWithCli(context.Background(), cli)

	mockKV := NewMockKV(t)

	cli.KV = mockKV

	mockKV.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(&clientv3.PutResponse{}, nil)

	err := etcd.Put("key", "value")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPut_Error(t *testing.T) {
	mockKV := NewMockKV(t)
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)
	expectedErr := errors.New("test error")
	etcd := newEtcdHandlerWithCli(ctx, cli)

	cli.KV = mockKV

	mockKV.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr)

	err := etcd.Put("key", "value")

	if err != expectedErr {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestWatch_Success(t *testing.T) {
	mockWatcher := NewMockWatcher(t)
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)
	etcd := newEtcdHandlerWithCli(ctx, cli)

	cli.Watcher = mockWatcher

	watchRespChan := make(chan clientv3.WatchResponse)

	go func() {
		watchRespChan <- clientv3.WatchResponse{
			Events: []*clientv3.Event{
				{
					Type: clientv3.EventTypePut,
					Kv: &mvccpb.KeyValue{
						Key:   []byte("key"),
						Value: []byte("value"),
					},
				},
				{
					Type: clientv3.EventTypePut,
					Kv: &mvccpb.KeyValue{
						Key:   []byte("key"),
						Value: []byte("anotherValue"),
					},
				},
			},
		}

		close(watchRespChan)
	}()

	mockWatcher.EXPECT().Watch(mock.Anything, mock.Anything).Return(watchRespChan)

	watchChan, cancel := etcd.Watch("key")

	expectedMap := make(map[int]string)
	expectedMap[0] = "value"
	expectedMap[1] = "anotherValue"

	for resp := range watchChan {
		for i, ev := range resp.Events {
			if expectedMap[i] != string(ev.Kv.Value) || string(ev.Kv.Key) != "key" {
				t.Errorf("Expected event at pos %d, to have key: %s and value: %s", i, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}

	cancel()
}

func newEtcdHandlerWithCli(ctx context.Context, cli *clientv3.Client) Etcd {
	return &EtcdHandler{
		conf: clientv3.Config{},
		cli:  cli,
		ctx:  ctx,
	}
}
