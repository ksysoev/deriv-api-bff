package config

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/spf13/viper"
)

type FileSource struct {
	mu            sync.RWMutex
	reader        *viper.Viper
	currentConfig []handlerfactory.Config
	path          string
	onChange      func([]handlerfactory.Config)
}

type configUpdates struct {
	Value any
	Found bool
}

func NewFileSource(path string) (*FileSource, error) {
	reader := viper.New()
	reader.SetConfigFile(path)

	return &FileSource{
		reader: reader,
	}, nil
}

func (fs *FileSource) WatchConfig(onChange func([]handlerfactory.Config)) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.onChange != nil {
		return fmt.Errorf("config watcher already set")
	}

	fs.onChange = onChange

	fs.reader.WatchConfig()
	fs.reader.OnConfigChange(fs.onFileChange)

	return nil
}

func (fs *FileSource) GetConfig() []handlerfactory.Config {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.currentConfig
}

func (fs *FileSource) LoadConfig() ([]handlerfactory.Config, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := fs.reader.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg []handlerfactory.Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	fs.currentConfig = cfg

	return cfg, nil
}

func (fs *FileSource) onFileChange(in fsnotify.Event) {
	slog.Debug(fmt.Sprintf("config file changed at %s", in.Name))

	cfg, err := fs.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config on change", slog.Any("error", err))
		return
	}

	fs.onChange(cfg)
}
