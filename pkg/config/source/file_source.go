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

// NewFileSource creates a new instance of FileSource with the given file path.
// It takes a single parameter path of type string which specifies the file path.
// It returns a pointer to a FileSource initialized with the provided path.
func NewFileSource(path string) *FileSource {
	return &FileSource{
		path: path,
	}
}

// LoadConfig loads the configuration from a file or directory specified by the FileSource path.
// It takes a context parameter which is currently unused.
// It returns a slice of handlerfactory.Config and an error.
// It returns an error if the file or directory cannot be accessed, or if the configuration cannot be read.
// If the path points to a directory, it reads all configuration files within the directory.
// If the path points to a regular file, it reads the configuration from that file.
// It returns an error if the path points to an unsupported file type.
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

// readDir reads all YAML files from the specified directory and returns their configurations.
// It takes a single parameter path of type string which is the directory path to read from.
// It returns a slice of handlerfactory.Config and an error.
// It returns an error if the directory cannot be read or if any file cannot be read.
// It skips directories and non-YAML files within the specified directory.
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

		cfg, err := readFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		data = append(data, cfg...)
	}

	return data, nil
}

// readFile reads a YAML file from the given path and decodes its content into a slice of handlerfactory.Config.
// It takes a single parameter, path, which is a string representing the file path.
// It returns a slice of handlerfactory.Config and an error.
// It returns an error if the file cannot be opened or if the content cannot be decoded.
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

// isYamlFile checks if the given file path has a YAML file extension.
// It takes a single parameter path of type string.
// It returns a boolean value: true if the file has a .yaml or .yml extension, otherwise false.
func isYamlFile(path string) bool {
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}
