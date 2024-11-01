package config

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAddConfigSource(t *testing.T) {
	cfg := &Config{}

	assert.Equal(t, 0, len(cfg.sources))

	cfg.addConfigSource(&FileSource{})

	assert.Equal(t, 1, len(cfg.sources))
}

func TestWatchConfig_Error(t *testing.T) {
	cfg := &Config{}

	cfg.addConfigSource(NewFileSource("/path/to/nofile"))

	err := cfg.WatchConfig(NewEvent[any](), "key")

	assert.Error(t, err)
	assert.Equal(t, "watch on key failed. Linked event has no handlers", err.Error())
}

func TestWatchConfig_Success(t *testing.T) {
	cfg := &Config{}
	event := NewEvent[any]()

	cfg.addConfigSource(NewFileSource("/path/to/nofile"))
	event.RegisterHandler(func(_ context.Context, a any) {
		print(a)
	})

	err := cfg.WatchConfig(event, "key")

	assert.NoError(t, err)
}

func TestWatchConfig_WithFileChanges(t *testing.T) {
	viper.Reset()

	cfg := &Config{}
	event := NewEvent[any]()
	configFile := createTempConfigFile(t, "")
	source := NewFileSource(configFile)
	isUpdated := new(atomic.Bool)

	cfg.addConfigSource(source)
	event.RegisterHandler(func(_ context.Context, a any) {
		callsMap, _ := a.(map[string]any)

		assert.Equal(t, 0, len(callsMap))
		isUpdated.CompareAndSwap(false, true)
	})

	err := source.Init(cfg)
	assert.NoError(t, err)

	err = cfg.WatchConfig(event, "api.calls")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(source.GetWatchKeys()))

	// now modify the file
	createTempConfigFile(t, validConfig)
	time.Sleep(1 * time.Second)

	assert.Equal(t, true, isUpdated.Load())
}
