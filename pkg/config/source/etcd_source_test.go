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

func TestEtcdSource_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	etcdHost, err := ctr.ClientEndpoints(ctx)
	require.NoError(t, err)

	t.Run("Fail to create etcd source with empty servers", func(t *testing.T) {
		_, err := NewEtcdSource(EtcdConfig{
			Servers: "",
			Prefix:  "",
		})
		assert.Error(t, err, "should fail to create etcd source")
	})

	t.Run("Fail to load config with empty servers", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: "",
			Prefix:  "test1::",
		})
		require.NoError(t, err)

		configs, err := source.LoadConfig(ctx)
		assert.Error(t, err, "failed to load config")
		assert.Nil(t, configs)
	})

	t.Run("Fail to put config with empty servers", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: "",
			Prefix:  "test1::",
		})
		require.NoError(t, err)

		err = source.PutConfig(ctx, []handlerfactory.Config{
			{
				Method: "Test",
			},
		})
		assert.Error(t, err, "failed to put config")
	})

	t.Run("Create etcd source with valid servers", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		assert.NoError(t, err, "failed to create etcd source")
		assert.NotNil(t, source)

		configs, err := source.LoadConfig(ctx)
		assert.NoError(t, err, "failed to load config")
		assert.NotNil(t, configs)
		assert.Len(t, configs, 0)
	})

	t.Run("Put and load valid config", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		require.NoError(t, err)

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

		configs, err := source.LoadConfig(ctx)
		assert.NoError(t, err, "failed to load config")
		assert.Equal(t, expected, configs)
	})

	t.Run("Fail to marshal invalid config", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		require.NoError(t, err)

		err = source.PutConfig(ctx, []handlerfactory.Config{
			{
				Method: "Test",
				Params: &validator.Config{"test": make(chan int)},
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
		})
		assert.Error(t, err, "failed to marshal config")
	})

	t.Run("Fail to load invalid JSON config", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		require.NoError(t, err)

		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints:   etcdHost,
			DialTimeout: defaultTimeoutSeconds * time.Second,
		})
		require.NoError(t, err)

		_, err = etcdClient.Put(ctx, "test2::Test", "invalid json")
		require.NoError(t, err)

		configs, err := source.LoadConfig(ctx)
		assert.Error(t, err, "failed to load config")
		assert.Nil(t, configs)
	})

	t.Run("Put and load another valid config", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		require.NoError(t, err)

		expected := []handlerfactory.Config{
			{
				Method: "Test2",
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
			{
				Method: "Test3",
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
		}

		err = source.PutConfig(ctx, expected)
		require.NoError(t, err)

		configs, err := source.LoadConfig(ctx)
		assert.NoError(t, err, "failed to load config")
		assert.Equal(t, expected, configs)
	})

	t.Run("Watch configuration", func(t *testing.T) {
		source, err := NewEtcdSource(EtcdConfig{
			Servers: strings.Join(etcdHost, ","),
			Prefix:  "test2::",
		})
		require.NoError(t, err)

		source.reducerInterval = 50 * time.Millisecond

		expected := []handlerfactory.Config{
			{
				Method: "Test2",
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
			{
				Method: "Test3",
				Backend: []*processor.Config{
					{
						Tmplt: map[string]any{"ping": "pong"},
					},
				},
			},
		}

		counter := 0
		done := make(chan struct{}, 1)
		onUpdate := func() {
			counter++
			done <- struct{}{}
		}

		ready := make(chan struct{})

		go func() {
			close(ready)
			source.Watch(ctx, onUpdate)
		}()

		select {
		case <-ready:
		case <-time.After(1 * time.Second):
			t.Fatal("failed to start watching")
		}

		err = source.PutConfig(ctx, expected)
		require.NoError(t, err)

		select {
		case <-done:
			assert.Equal(t, 1, counter)
		case <-time.After(1 * time.Second):
			t.Fatal("onUpdate was not triggered within the expected time")
		}
	})

}
func TestMakeReducer_Succcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	triggered := make(chan struct{}, 1)
	onUpdate := func() {
		triggered <- struct{}{}
	}

	reducer := makeReducer(ctx, onUpdate, 1*time.Microsecond)

	reducer()

	select {
	case <-triggered:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("onUpdate was not triggered within the expected time")
	}
}

func TestMakeReducer_NotTriggerBeforeInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	triggered := make(chan struct{}, 1)
	onUpdate := func() {
		triggered <- struct{}{}
	}

	reducer := makeReducer(ctx, onUpdate, 10*time.Millisecond)

	reducer()

	select {
	case <-triggered:
		t.Fatal("onUpdate was triggered too early")
	case <-time.After(1 * time.Millisecond): // Success
	}
}

func TestMakeReducer_OnlyOneEventPerInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	triggered := make(chan struct{}, 3)
	counter := 0
	onUpdate := func() {
		counter++
		triggered <- struct{}{}
	}

	reducer := makeReducer(ctx, onUpdate, 1*time.Millisecond)

	reducer()
	reducer()
	reducer()

	select {
	case <-triggered:
		assert.Equal(t, 1, counter)
	case <-time.After(10 * time.Millisecond):
		t.Fatal("onUpdate was not triggered within the expected time")
	}
}

func TestMakeReducer_ContextCancelation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	triggered := make(chan struct{}, 1)
	onUpdate := func() {
		triggered <- struct{}{}
	}

	reducer := makeReducer(ctx, onUpdate, 1*time.Millisecond)

	reducer()
	cancel()

	select {
	case <-triggered:
		t.Fatal("onUpdate wasn't triggered after context cancel")
	case <-time.After(10 * time.Millisecond): // Success
	}
}
