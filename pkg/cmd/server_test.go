package cmd

import (
	"context"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/core/validator"
	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	cfg := &config.Config{
		Server: api.Config{
			Listen: ":0",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServer(ctx, cfg)

	assert.NoError(t, err)
}

func TestRunServer_Error(t *testing.T) {
	cfg := &config.Config{
		Server: api.Config{
			Listen: ":0",
		},
		API: config.CallsConfig{
			Calls: []config.CallConfig{
				{
					Method: "GET",
					Params: validator.Config{
						"param": &validator.FieldSchema{
							Type: "InvalidType",
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServer(ctx, cfg)

	assert.Error(t, err)
}
