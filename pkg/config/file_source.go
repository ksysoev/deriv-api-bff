package config

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type FileSource struct {
	mu                *sync.RWMutex
	currentConfig     *Config
	watchKeyPrefixSet *map[string]struct{}
	configFilePath    string
}

func (fileSource *FileSource) Init() error {
	viper.SetConfigFile(fileSource.configFilePath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.OnConfigChange(fileSource.onFileChange)
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	slog.Debug("Config loaded", slog.Any("config", cfg))

	return nil
}

func (fileSource *FileSource) GetConfigurations() (*Config, error) {
	fileSource.mu.RLock()
	defer fileSource.mu.RUnlock()

	return fileSource.currentConfig, nil
}

func (fileSource *FileSource) WatchConfig(configKey string) {
	fileSource.mu.Lock()
	defer fileSource.mu.Unlock()

	keyPrefixes := *fileSource.watchKeyPrefixSet

	keyPrefixes[configKey] = struct{}{}
}

func (fileSource *FileSource) GetPriority() Priority { return P1 }

func (fileSource *FileSource) Name() string { return "VIPER_FILE_SOURCE" }

func (fileSource *FileSource) Close() error { return nil }

func NewFileSource(filePath string) *FileSource {
	prefixes := make(map[string]struct{})

	return &FileSource{
		configFilePath:    filePath,
		currentConfig:     &Config{},
		mu:                new(sync.RWMutex),
		watchKeyPrefixSet: &prefixes,
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

	diffs := Compare(*(fileSource.currentConfig), *newConfig, "")

	for k := range *fileSource.watchKeyPrefixSet {
		for _, diff := range diffs {
			if strings.HasPrefix(diff, k) {
				fileSource.currentConfig = newConfig
				// TODO: populate events from here and
				break
			}
		}
	}
}
