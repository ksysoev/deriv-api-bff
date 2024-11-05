package source

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdSource_Int(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	etcdHost, err := ctr.ClientEndpoints(ctx)
	require.NoError(t, err)

	_, err = NewEtcdSource(EtcdConfig{
		Servers: "",
		Prefix:  "",
	})
	assert.Error(t, err, "should fail to create etcd source")

	source, err := NewEtcdSource(EtcdConfig{
		Servers: "",
		Prefix:  "test1::",
	})
	require.NoError(t, err)

	configs, err := source.LoadConfig(ctx)

	assert.Error(t, err, "failed to load config")
	assert.Nil(t, configs)

	err = source.PutConfig(ctx, []handlerfactory.Config{
		{
			Method: "Test",
		},
	})
	assert.Error(t, err, "failed to put config")

	source, err = NewEtcdSource(EtcdConfig{
		Servers: strings.Join(etcdHost, ","),
		Prefix:  "test2::",
	})

	assert.NoError(t, err, "failed to create etcd source")
	assert.NotNil(t, source)

	configs, err = source.LoadConfig(ctx)

	assert.NoError(t, err, "failed to load config")
	assert.NotNil(t, configs)
	assert.Len(t, configs, 0)

	expected := []handlerfactory.Config{
		{
			Method: "Test",
			Backend: []*processor.Config{
				{
					Tmplt: map[string]any{"ping": "pong"},
				},
			},
		},
		{
			Method: "Test2",
			Backend: []*processor.Config{
				{
					Tmplt: map[string]any{"ping": "pong"},
				},
			},
		},
	}

	err = source.PutConfig(ctx, expected)
	require.NoError(t, err)

	configs, err = source.LoadConfig(ctx)

	assert.NoError(t, err, "failed to load config")
	assert.Equal(t, expected, configs)

	err = source.PutConfig(ctx, []handlerfactory.Config{
		{
			Method: "Test",
			Params: validator.Config{"test": make(chan int)},
			Backend: []*processor.Config{
				{
					Tmplt: map[string]any{"ping": "pong"},
				},
			},
		},
	})
	assert.Error(t, err, "failed to marshal config")

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdHost,
		DialTimeout: defaultTimeoutSeconds * time.Second,
	})
	require.NoError(t, err)

	_, err = etcdClient.Put(ctx, "test2::Test", "invalid json")
	require.NoError(t, err)

	configs, err = source.LoadConfig(ctx)
	assert.Error(t, err, "failed to load config")
	assert.Nil(t, configs)
}
