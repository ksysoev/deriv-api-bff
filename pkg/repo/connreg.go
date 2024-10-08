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

// NewConnectionRegistry creates and returns a new instance of ConnectionRegistry.
// It initializes the connections map to store connection objects.
// It returns a pointer to the newly created ConnectionRegistry.
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*core.Conn),
	}
}

// GetConnection retrieves an existing connection or creates a new one if it doesn't exist.
// It takes a clientConn of type wasabi.Connection.
// It returns a pointer to a core.Conn.
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

// removeConnection removes a connection from the ConnectionRegistry by its ID.
// It takes one parameter: id of type string, which is the identifier of the connection to be removed.
// It does not return any values.
func (c *ConnectionRegistry) removeConnection(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, id)
}
