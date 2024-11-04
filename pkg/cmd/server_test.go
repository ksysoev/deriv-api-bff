package cmd

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config/source"
	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	callsPath := createTempConfigFile(t, callsConfig)

	cfg := &Config{
		Server: api.Config{
			Listen: ":0",
		},
		APISource: source.Config{
			Path: callsPath,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServer(ctx, cfg)

	assert.NoError(t, err)
}

func TestRunServer_Error(t *testing.T) {
	cfg := &Config{
		Server: api.Config{
			Listen: ":0",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServer(ctx, cfg)

	assert.Error(t, err)
}

func TestRunServer_FailToCreateSources(t *testing.T) {
	cfg := &Config{
		Server: api.Config{
			Listen: ":0",
		},
		APISource: source.Config{
			Etcd: source.EtcdConfig{
				Servers: "localhost:2379",
				Prefix:  "",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServer(ctx, cfg)

	assert.Error(t, err)
}
