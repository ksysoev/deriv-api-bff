package config

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type FileSource struct {
	mu             *sync.RWMutex
	currentConfig  *Config
	watchKeys      map[string]*Event[any]
	configFilePath string
}

type configUpdates struct {
	Value any
	Found bool
}

func (fileSource *FileSource) Init(cfg *Config) error {
	fileSource.mu.Lock()
	defer fileSource.mu.Unlock()

	viper.SetConfigFile(fileSource.configFilePath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.OnConfigChange(fileSource.onFileChange)
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))
	cfg.addConfigSource(fileSource)

	fileSource.currentConfig = cfg

	return nil
}

func (fileSource *FileSource) GetConfigurations() (*Config, error) {
	fileSource.mu.RLock()
	defer fileSource.mu.RUnlock()

	return fileSource.currentConfig, nil
}

func (fileSource *FileSource) WatchConfig(event *Event[any], key string) error {
	fileSource.mu.Lock()
	defer fileSource.mu.Unlock()

	if len(event.subscribers) == 0 {
		return fmt.Errorf("watch on %s failed. Linked event has no handlers", key)
	}

	fileSource.watchKeys[key] = event

	return nil
}

func (fileSource *FileSource) GetWatchKeys() map[string]*Event[any] {
	fileSource.mu.RLock()
	defer fileSource.mu.RUnlock()

	return fileSource.watchKeys
}

func (fileSource *FileSource) GetPriority() Priority { return P1 }

func (fileSource *FileSource) Name() string { return "VIPER_FILE_SOURCE" }

func (fileSource *FileSource) Close() error { return nil }

func NewFileSource(filePath string) *FileSource {
	return &FileSource{
		configFilePath: filePath,
		currentConfig:  &Config{},
		mu:             new(sync.RWMutex),
		watchKeys:      make(map[string]*Event[any]),
	}
}

func (fileSource *FileSource) onFileChange(in fsnotify.Event) {
	slog.Debug(fmt.Sprintf("config file changed at %s", in.Name))

	fileSource.mu.Lock()
	defer fileSource.mu.Unlock()

	newConfig := &Config{}

	if err := viper.Unmarshal(newConfig); err != nil {
		slog.Error(fmt.Sprintf("Error while unmarshal on update from file %s: %v", in.Name, err))
		return
	}

	fileSource.currentConfig = newConfig

	keys := make([]string, len(fileSource.watchKeys))
	i := 0

	for k := range fileSource.watchKeys {
		keys[i] = k
		i++
	}

	watchedUpdates := findKeyValue(viper.AllSettings(), keys)

	for idx, key := range keys {
		update := watchedUpdates[idx]
		if update.Found {
			event := fileSource.watchKeys[key]

			event.Notify(context.Background(), update.Value)
		}
	}
}

func findKeyValue(m map[string]any, keyList []string) []configUpdates {
	results := make([]configUpdates, len(keyList))

	curr := m

	for i, k := range keyList {
		children := strings.Split(k, ".")

		for _, child := range children {
			v, exists := curr[child]
			if exists {
				results[i] = configUpdates{Value: v, Found: true}

				if nested, ok := v.(map[string]any); ok {
					curr = nested
				} else {
					break
				}
			} else {
				results[i] = configUpdates{Value: nil, Found: false}
			}
		}
	}

	return results
}
