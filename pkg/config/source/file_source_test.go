package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSource(t *testing.T) {
	path := "/some/path/to/config"
	fs := NewFileSource(path)

	assert.NotNil(t, fs, "FileSource should not be nil")
	assert.Equal(t, path, fs.path, "FileSource path should match the input path")
}
func TestIsYamlFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"YAML file with .yaml extension", "/some/path/to/config.yaml", true},
		{"YAML file with .yml extension", "/some/path/to/config.yml", true},
		{"Non-YAML file with .json extension", "/some/path/to/config.json", false},
		{"Non-YAML file with no extension", "/some/path/to/config", false},
		{"Non-YAML file with .txt extension", "/some/path/to/config.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isYamlFile(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		expected    []handlerfactory.Config
		expectError bool
	}{
		{
			name: "Valid YAML file",
			fileContent: `
- method: config1
- method: config2
`,
			expected: []handlerfactory.Config{
				{Method: "config1"},
				{Method: "config2"},
			},
			expectError: false,
		},
		{
			name:        "Invalid YAML file",
			fileContent: `invalid yaml content`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Empty file",
			fileContent: ``,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "testfile*.yaml")
			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			// Write the test content to the temporary file
			_, err = tmpFile.WriteString(tt.fileContent)
			assert.NoError(t, err)

			// Close the file to flush the content
			err = tmpFile.Close()
			assert.NoError(t, err)

			// Call the readFile function
			result, err := readFile(tmpFile.Name())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReadFile_NotExistFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile*.yaml")
	require.NoError(t, err)
	tmpFile.Close()

	result, err := readFile(tmpFile.Name() + "/nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReadDir(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		expected    []handlerfactory.Config
		expectError bool
	}{
		{
			name: "Directory with valid YAML files",
			files: map[string]string{
				"config1.yaml": `
- method: config1
`,
				"config2.yaml": `
- method: config2
`,
			},
			expected: []handlerfactory.Config{
				{Method: "config1"},
				{Method: "config2"},
			},
			expectError: false,
		},
		{
			name: "Directory with invalid YAML file",
			files: map[string]string{
				"invalid.yaml": `invalid yaml content`,
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Directory with non-YAML files",
			files: map[string]string{
				"config.json": `{"method": "config1"}`,
				"config.txt":  `some text content`,
			},
			expected:    nil,
			expectError: false,
		},
		{
			name:        "Empty directory",
			files:       map[string]string{},
			expected:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "testdir")
			require.NoError(t, err)

			defer os.RemoveAll(tmpDir)

			// Create the test files in the temporary directory
			for name, content := range tt.files {
				err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o600)
				require.NoError(t, err)
			}

			// Call the readDir function
			result, err := readDir(tmpDir)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReadDir_NotExistDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	result, err := readDir(tmpDir + "/nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expected    []handlerfactory.Config
		expectError bool
	}{
		{
			name: "Load config from valid YAML file",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpFile, err := os.CreateTemp("", "testfile*.yaml")
				assert.NoError(t, err)
				defer tmpFile.Close()

				content := `
- method: config1
- method: config2
`
				_, err = tmpFile.WriteString(content)
				assert.NoError(t, err)

				return tmpFile.Name()
			},
			expected: []handlerfactory.Config{
				{Method: "config1"},
				{Method: "config2"},
			},
			expectError: false,
		},
		{
			name: "Load config from directory with valid YAML files",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir, err := os.MkdirTemp("", "testdir")
				assert.NoError(t, err)

				files := map[string]string{
					"config1.yaml": `
- method: config1
`,
					"config2.yaml": `
- method: config2
`,
				}

				for name, content := range files {
					err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o600)
					assert.NoError(t, err)
				}

				return tmpDir
			},
			expected: []handlerfactory.Config{
				{Method: "config1"},
				{Method: "config2"},
			},
			expectError: false,
		},
		{
			name: "Load config from invalid YAML file",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpFile, err := os.CreateTemp("", "testfile*.yaml")
				assert.NoError(t, err)
				defer tmpFile.Close()

				content := `invalid yaml content`
				_, err = tmpFile.WriteString(content)
				assert.NoError(t, err)

				return tmpFile.Name()
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Load config from unsupported file type",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpFile, err := os.CreateTemp("", "testfile*.txt")
				assert.NoError(t, err)
				defer tmpFile.Close()

				content := `some text content`
				_, err = tmpFile.WriteString(content)
				assert.NoError(t, err)

				return tmpFile.Name()
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Load config from empty directory",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir, err := os.MkdirTemp("", "testdir")
				assert.NoError(t, err)
				return tmpDir
			},
			expected:    nil,
			expectError: false,
		},
		{
			name: "Invalid Path",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir, err := os.MkdirTemp("", "testdir")
				assert.NoError(t, err)
				return tmpDir + "/nonexistent"
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Skip directories",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir, err := os.MkdirTemp("", "testdir")
				assert.NoError(t, err)

				err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0o700)
				assert.NoError(t, err)

				return tmpDir
			},
			expected:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			fs := NewFileSource(path)

			result, err := fs.LoadConfig(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
