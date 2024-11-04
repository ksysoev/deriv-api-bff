package source

import "testing"

func TestCreateOptions(t *testing.T) {
	tests := []struct {
		cfg     *Config
		name    string
		wantErr bool
	}{
		{
			name: "Valid Path Config",
			cfg: &Config{
				Path: "/valid/path",
			},
			wantErr: false,
		},
		{
			name: "Valid Etcd Config",
			cfg: &Config{
				Etcd: EtcdConfig{
					Servers: "http://localhost:2379",
					Prefix:  "testApiPrefix::",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Path and Etcd Config",
			cfg: &Config{
				Path: "/valid/path",
				Etcd: EtcdConfig{
					Servers: "http://localhost:2379",
					Prefix:  "testApiPrefix::",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Etcd Config",
			cfg: &Config{
				Etcd: EtcdConfig{
					Servers: "http://localhost:2379",
				},
			},
			wantErr: true,
		},
		{
			name:    "Empty Config",
			cfg:     &Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateOptions(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
