package source

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
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

	source, err := NewEtcdSource(EtcdConfig{
		Servers: strings.Join(etcdHost, ","),
		Prefix:  "test::",
	})

	assert.NoError(t, err, "failed to create etcd source")
	assert.NotNil(t, source)

	configs, err := source.LoadConfig(ctx)

	assert.NoError(t, err, "failed to load config")
	assert.NotNil(t, configs)
	assert.Len(t, configs, 0)

	expected := []handlerfactory.Config{
		{
			Method: "Test",
		},
		{
			Method: "Test2",
		},
	}

	err = source.PutConfig(ctx, expected)
	require.NoError(t, err)

	configs, err = source.LoadConfig(ctx)

	assert.NoError(t, err, "failed to load config")
	assert.Equal(t, expected, configs)
}
