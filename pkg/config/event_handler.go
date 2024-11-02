package config

import "context"

type EventHandler[T any] func(context.Context, T)
