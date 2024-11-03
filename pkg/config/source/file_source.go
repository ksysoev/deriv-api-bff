package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"gopkg.in/yaml.v3"
)

type FileSource struct {
	path string
	mu   sync.RWMutex
}

func NewFileSource(path string) *FileSource {
	return &FileSource{
		path: path,
	}
}

func (fs *FileSource) LoadConfig(_ context.Context) ([]handlerfactory.Config, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fi, err := os.Stat(fs.path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	var cfg []handlerfactory.Config

	switch mode := fi.Mode(); {
	case mode.IsDir():
		cfg, err = readDir(fs.path)
	case mode.IsRegular():
		cfg, err = readFile(fs.path)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}

func readDir(path string) ([]handlerfactory.Config, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var data []handlerfactory.Config

	for _, file := range files {
		// TODO: Shall we do recursive reading?
		if file.IsDir() {
			continue
		}

		if !isYamlFile(file.Name()) {
			continue
		}

		cfg, err := readFile(file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		data = append(data, cfg...)
	}

	return data, nil
}

func readFile(path string) ([]handlerfactory.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	y := yaml.NewDecoder(file)

	var data []handlerfactory.Config

	if err = y.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	return data, nil
}

func isYamlFile(path string) bool {
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}
