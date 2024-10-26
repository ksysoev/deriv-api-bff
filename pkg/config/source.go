package config

// defines the source of the config and/or config changes“
type Source interface {
	Init() error

	GetConfigurations() (*Config, error)

	GetConfigurationByKey(string) ([]byte, error)

	WatchConfig(string)

	GetPriority() Priority

	Name() string

	Close()
}

type Priority int8

const (
	P0 Priority = iota
	P1
	P2
)

type EventType int8

const (
	INSERT EventType = iota + 1
	UPDATE
	DELETE
)
