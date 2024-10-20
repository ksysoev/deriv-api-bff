package repo

import (
	"context"
	"testing"

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
	etcd := NewMockEtcd(t)
	cfg := EtcdConfig{
		Servers:            []string{"localhost:7000"},
		DialTimeoutSeconds: 1,
	}

	etcd.EXPECT().Client(cfg).Return(
		&clientv3.Client{
			Username: "test",
			Password: "pass",
		}, nil)

	cli, err := etcd.Client(cfg)

	if err != nil {
		t.Errorf("Unexpected err: %s", err)
	}

	assert.Equal(t, "test", cli.Username)
	assert.Equal(t, "pass", cli.Password)
}

func TestPut(t *testing.T) {
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
