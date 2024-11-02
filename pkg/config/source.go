package config

// defines the source of the config and/or config changes“
type Source interface {
	Init(*Config) error

	GetConfigurations() (*Config, error)

	WatchConfig(*Event[any], string) error

	GetPriority() Priority

	Name() string

	Close() error
}

type Priority int8

const (
	P0 Priority = iota
	P1
	P2
)
