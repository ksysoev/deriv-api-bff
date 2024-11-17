package config

import (
	context "context"
	"fmt"
	"os"
	"testing"
	"time"

	handlerfactory "github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/ksysoev/deriv-api-bff/pkg/core/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var validConfig = []handlerfactory.Config{
	{
		Method: "handler1",
		Backend: []*processor.Config{
			{
				Request: map[string]any{"ping": 1},
			},
		},
	},
	{
		Method: "handler2",
		Backend: []*processor.Config{
			{
				Request: map[string]any{"ping": 1},
			},
		},
	},
}

func TestNewService(t *testing.T) {
	mockBFFService := NewMockBFFService(t)
	mockLocalSource := NewMockLocalSource(t)
	mockRemoteSource := NewMockRemoteSource(t)

	tests := []struct {
		name    string
		bff     BFFService
		opts    []Option
		wantErr bool
	}{
		{
			name:    "No sources provided",
			bff:     mockBFFService,
			opts:    []Option{},
			wantErr: true,
		},
		{
			name:    "Only local source provided",
			bff:     mockBFFService,
			opts:    []Option{WithLocalSource(mockLocalSource)},
			wantErr: false,
		},
		{
			name:    "Only remote source provided",
			bff:     mockBFFService,
			opts:    []Option{WithRemoteSource(mockRemoteSource)},
			wantErr: false,
		},
		{
			name:    "Both local and remote sources provided",
			bff:     mockBFFService,
			opts:    []Option{WithLocalSource(mockLocalSource), WithRemoteSource(mockRemoteSource)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(tt.bff, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestService_LoadHandlers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		err        error
		loadConfig func() ([]handlerfactory.Config, error)
		name       string
		cfg        []handlerfactory.Config
		local      bool
		wantErr    bool
	}{
		{
			name:  "Load from local source",
			local: true,
			cfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			err:     nil,
			wantErr: false,
		},
		{
			name:  "Load from remote source",
			local: false,
			cfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			err:     nil,
			wantErr: false,
		},
		{
			name:    "Load from local source with error",
			local:   true,
			cfg:     nil,
			err:     fmt.Errorf("error loading config"),
			wantErr: true,
		},
		{
			name:    "Load from remote source with error",
			local:   false,
			cfg:     nil,
			err:     fmt.Errorf("error loading config"),
			wantErr: true,
		},
		{
			name:  "Duplicate handler names",
			local: false,
			cfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			err:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			mockLocalSource := NewMockLocalSource(t)
			mockRemoteSource := NewMockRemoteSource(t)

			opts := make([]Option, 0)

			if tt.local {
				opts = append(opts, WithLocalSource(mockLocalSource))
				mockLocalSource.EXPECT().LoadConfig(ctx).Return(tt.cfg, tt.err)
			} else {
				opts = append(opts, WithRemoteSource(mockRemoteSource))
				mockRemoteSource.EXPECT().LoadConfig(ctx).Return(tt.cfg, tt.err)
			}

			svc, err := New(mockBFFService, opts...)
			require.NoError(t, err)

			if !tt.wantErr {
				mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return()
			}

			err = svc.LoadHandlers(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc.curCfg)
			}
		})
	}
}

func TestService_PutConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		loadErr error
		putErr  error
		name    string
		curCfg  []handlerfactory.Config
		local   bool
		remote  bool
		wantErr bool
	}{
		{
			name:    "Local source missing",
			local:   false,
			remote:  true,
			wantErr: true,
		},
		{
			name:    "Remote source missing",
			local:   true,
			remote:  false,
			wantErr: true,
		},
		{
			name:    "Load handlers error",
			local:   true,
			remote:  true,
			loadErr: fmt.Errorf("load handlers error"),
			wantErr: true,
		},
		{
			name:   "Put config error",
			local:  true,
			remote: true,
			curCfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			putErr:  fmt.Errorf("put config error"),
			wantErr: true,
		},
		{
			name:   "Successful put config",
			local:  true,
			remote: true,
			curCfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			mockLocalSource := NewMockLocalSource(t)
			mockRemoteSource := NewMockRemoteSource(t)

			opts := make([]Option, 0)

			if tt.local {
				opts = append(opts, WithLocalSource(mockLocalSource))
				mockLocalSource.EXPECT().LoadConfig(ctx).Return(tt.curCfg, tt.loadErr).Maybe()
			}

			if tt.remote {
				opts = append(opts, WithRemoteSource(mockRemoteSource))
				mockRemoteSource.EXPECT().PutConfig(ctx, tt.curCfg).Return(tt.putErr).Maybe()
			}

			svc, err := New(mockBFFService, opts...)
			require.NoError(t, err)

			if tt.local && tt.remote {
				mockLocalSource.EXPECT().LoadConfig(ctx).Return(tt.curCfg, tt.loadErr)

				if tt.loadErr == nil {
					mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return()
				}
			}

			err = svc.PutConfig(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateHandlers(t *testing.T) {
	tests := []struct {
		name    string
		cfg     []handlerfactory.Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			cfg:     validConfig,
			wantErr: false,
		},
		{
			name: "Duplicate handler names",
			cfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Handler creation error",
			cfg: []handlerfactory.Config{
				{Method: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, err := createHandlers(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, handlers)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handlers)
				assert.Equal(t, len(tt.cfg), len(handlers))
			}
		})
	}
}
func TestService_ProcessConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     []handlerfactory.Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			cfg:     validConfig,
			wantErr: false,
		},
		{
			name: "Duplicate handler names",
			cfg: []handlerfactory.Config{
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
				{
					Method: "handler1",
					Backend: []*processor.Config{
						{
							Request: map[string]any{"ping": 1},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Handler creation error",
			cfg: []handlerfactory.Config{
				{Method: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			svc := &Service{
				bff: mockBFFService,
			}

			if !tt.wantErr {
				mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return()
			}

			err := svc.processConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.cfg, svc.curCfg)
			}
		})
	}
}

func TestService_Start(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		loadErr    error
		processErr error
		name       string
		local      bool
		remote     bool
		wantErr    bool
	}{
		{
			name:    "Load from local source",
			local:   true,
			remote:  false,
			loadErr: nil,
			wantErr: false,
		},
		{
			name:    "Load from remote source",
			local:   false,
			remote:  true,
			loadErr: nil,
			wantErr: false,
		},
		{
			name:    "Load from local source with error",
			local:   true,
			remote:  false,
			loadErr: fmt.Errorf("error loading config"),
			wantErr: true,
		},
		{
			name:    "Load from remote source with error",
			local:   false,
			remote:  true,
			loadErr: fmt.Errorf("error loading config"),
			wantErr: true,
		},
		{
			name:       "Process config error",
			local:      true,
			remote:     false,
			loadErr:    nil,
			processErr: fmt.Errorf("error processing config"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			mockLocalSource := NewMockLocalSource(t)
			mockRemoteSource := NewMockRemoteSource(t)

			opts := make([]Option, 0)

			var expectedCfg []handlerfactory.Config
			if tt.processErr == nil {
				expectedCfg = []handlerfactory.Config{}
			} else {
				expectedCfg = []handlerfactory.Config{{Method: "handler1"}}
			}

			if tt.local {
				opts = append(opts, WithLocalSource(mockLocalSource))
				mockLocalSource.EXPECT().LoadConfig(mock.Anything).Return(expectedCfg, tt.loadErr)
			}

			if tt.remote {
				opts = append(opts, WithRemoteSource(mockRemoteSource))
				mockRemoteSource.EXPECT().LoadConfig(mock.Anything).Return(expectedCfg, tt.loadErr)
			}

			svc, err := New(mockBFFService, opts...)
			require.NoError(t, err)

			if tt.loadErr == nil && tt.processErr == nil {
				mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return()
			}

			watched := make(chan struct{})

			if tt.loadErr == nil && tt.processErr == nil && tt.remote {
				mockRemoteSource.EXPECT().Watch(mock.Anything, mock.MatchedBy(func(_ any) bool {
					close(watched)
					return true
				})).Return()
			} else {
				close(watched)
			}

			err = svc.Start(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			svc.Stop()

			select {
			case <-watched:
			case <-time.After(1 * time.Second):
				t.Error("Expected Watch to be called")
			}
		})
	}
}

func TestService_onUpdate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		loadErr    error
		processErr error
		name       string
	}{
		{
			name:       "Successful update",
			loadErr:    nil,
			processErr: nil,
		},
		{
			name:       "Load config error",
			loadErr:    fmt.Errorf("error loading config"),
			processErr: nil,
		},
		{
			name:       "Process config error",
			loadErr:    nil,
			processErr: fmt.Errorf("error processing config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			mockRemoteSource := NewMockRemoteSource(t)

			svc := &Service{
				bff:    mockBFFService,
				remote: mockRemoteSource,
			}

			var expectedCfg []handlerfactory.Config
			if tt.processErr == nil {
				expectedCfg = []handlerfactory.Config{}
			} else {
				expectedCfg = []handlerfactory.Config{{Method: "handler1"}}
			}

			mockRemoteSource.EXPECT().LoadConfig(ctx).Return(expectedCfg, tt.loadErr)

			if tt.loadErr == nil {
				mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return().Maybe()
			}

			if tt.loadErr == nil && tt.processErr == nil {
				mockBFFService.EXPECT().UpdateHandlers(mock.Anything).Return()
			}

			onUpdateFunc := svc.onUpdate(ctx)
			onUpdateFunc()
		})
	}
}
func TestService_WriteConfig(t *testing.T) {
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	tests := []struct {
		name     string
		localCfg []handlerfactory.Config
		localErr error
		noRemote bool
		wantErr  bool
	}{
		{
			name:     "Successful write config",
			localCfg: validConfig,
			wantErr:  false,
			noRemote: false,
		},
		{
			name:     "Remote source missing",
			wantErr:  true,
			noRemote: true,
		},
		{
			name:     "Load config error",
			localErr: fmt.Errorf("error loading config"),
			wantErr:  true,
			noRemote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBFFService := NewMockBFFService(t)
			mockRemoteSource := NewMockRemoteSource(t)

			if tt.noRemote {
				mockRemoteSource = nil
			}

			svc := &Service{
				bff:    mockBFFService,
				remote: mockRemoteSource,
			}

			if tt.localErr != nil || tt.localCfg != nil {
				mockRemoteSource.EXPECT().LoadConfig(ctx).Return(tt.localCfg, tt.localErr)
			}

			filePath := fmt.Sprintf("%s/%s", tmpDir, tt.name)

			err := svc.WriteConfig(ctx, filePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
