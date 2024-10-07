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
// The returned ConnectionRegistry is initialized with an empty map to store connections.
// This function is typically used to initialize a ConnectionRegistry before adding connections to it.
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*core.Conn),
	}
}

// GetConnection retrieves an existing connection associated with the given client connection
// or creates a new one if it does not exist. It ensures thread-safe access to the connection
// registry. The returned connection is an instance of *core.Conn.
//
// Parameters:
//   - clientConn: The client connection for which the corresponding core.Conn is requested.
//
// Returns:
//   - *core.Conn: The connection associated with the provided client connection.
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

// removeConnection removes a connection from the registry by its unique identifier.
// This method is thread-safe and ensures that the connection is properly removed
// from the internal storage. Use this function to clean up connections that are
// no longer needed or have been closed.
//
// Parameters:
//
//	id - The unique identifier of the connection to be removed.
func (c *ConnectionRegistry) removeConnection(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, id)
}
