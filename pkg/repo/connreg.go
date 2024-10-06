package repo

import (
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/wasabi"
)

type ConnectionRegistry struct {
	connections map[string]*core.Conn
	mu          sync.RWMutex
}

func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*core.Conn),
	}
}

func (c *ConnectionRegistry) GetConnection(clientConn wasabi.Connection) *core.Conn {
	c.mu.RLock()
	conn, ok := c.connections[clientConn.ID()]
	c.mu.RUnlock()

	if ok {
		return conn
	}

	return c.AddConnection(clientConn)
}

func (c *ConnectionRegistry) AddConnection(clientConn wasabi.Connection) *core.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn := core.NewConnection(clientConn, c.RemoveConnection)
	c.connections[clientConn.ID()] = conn

	return conn
}

func (c *ConnectionRegistry) RemoveConnection(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, id)
}
