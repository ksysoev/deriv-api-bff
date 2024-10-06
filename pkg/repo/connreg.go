package repo

import (
	"sync"

	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/wasabi"
)

type ConnectionRegistry struct {
	connections map[string]*core.Conn
	mu          sync.Mutex
}

func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*core.Conn),
	}
}

func (c *ConnectionRegistry) GetConnection(clientConn wasabi.Connection) *core.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.connections[clientConn.ID()]; ok {
		return conn
	}

	conn := core.NewConnection(clientConn, c.removeConnection)
	c.connections[clientConn.ID()] = conn

	return conn
}

func (c *ConnectionRegistry) removeConnection(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, id)
}
