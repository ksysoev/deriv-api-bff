package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/etcd/clientv3"
)

func TestNew(t *testing.T) {
	etcd := NewEtcdHandler(EtcdConfig{
		Servers:            []string{"localhost:7000"},
		DialTimeoutSeconds: 1,
	})

	assert.NotEmpty(t, etcd)
	assert.Implements(t, (*Etcd)(nil), etcd)
}

func TestClient(t *testing.T) {
	cfg := EtcdConfig{
		Servers:            []string{"localhost:7000"},
		DialTimeoutSeconds: 1,
	}
	etcd := NewEtcdHandler(cfg)

	cli, err := etcd.Client()

	if err != nil {
		t.Errorf("Unexpected err: %s", err)
	}

	assert.ElementsMatch(t, []string{"localhost:7000"}, cli.Endpoints())
}

func TestPut_Success(t *testing.T) {
	etcd := NewEtcdHandler(EtcdConfig{})
	mockKV := NewMockKV(t)
	mockLease := NewMockLease(t)
	mockWatcher := NewMockWatcher(t)
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)

	cli.KV = mockKV
	cli.Lease = mockLease
	cli.Watcher = mockWatcher

	mockKV.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(&clientv3.PutResponse{}, nil)
	mockWatcher.EXPECT().Close().Return(nil)
	mockLease.EXPECT().Close().Return(nil)

	err := etcd.Put(ctx, cli, "key", "value")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPut_Error(t *testing.T) {
	etcd := NewEtcdHandler(EtcdConfig{})
	mockKV := NewMockKV(t)
	mockLease := NewMockLease(t)
	mockWatcher := NewMockWatcher(t)
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)
	expectedErr := errors.New("test error")

	cli.KV = mockKV
	cli.Lease = mockLease
	cli.Watcher = mockWatcher

	mockKV.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr)
	mockWatcher.EXPECT().Close().Return(nil)
	mockLease.EXPECT().Close().Return(nil)

	err := etcd.Put(ctx, cli, "key", "value")

	if err != expectedErr {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestWatch_Success(t *testing.T) {
	etcd := NewEtcdHandler(EtcdConfig{})
	mockKV := NewMockKV(t)
	mockLease := NewMockLease(t)
	mockWatcher := NewMockWatcher(t)
	ctx := context.Background()
	cli := clientv3.NewCtxClient(ctx)

	cli.KV = mockKV
	cli.Lease = mockLease
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
	mockWatcher.EXPECT().Close().Return(nil)
	mockLease.EXPECT().Close().Return(nil)

	watchChan, cancel := etcd.Watch(ctx, cli, "key")

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
