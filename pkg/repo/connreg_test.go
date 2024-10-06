package repo

import (
	"testing"

	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionRegistry(t *testing.T) {
	registry := NewConnectionRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.connections)
}

func TestGetConnection(t *testing.T) {
	registry := NewConnectionRegistry()
	clientConn := mocks.NewMockConnection(t)

	clientConn.EXPECT().ID().Return("test-conn")

	conn := registry.GetConnection(clientConn)
	assert.NotNil(t, conn)
	assert.Equal(t, clientConn.ID(), conn.ID())

	// Test if the same connection is returned for the same client connection
	sameConn := registry.GetConnection(clientConn)
	assert.Equal(t, conn, sameConn)
}

func TestRemoveConnection(t *testing.T) {
	registry := NewConnectionRegistry()
	clientConn := mocks.NewMockConnection(t)

	clientConn.EXPECT().ID().Return("test-conn")

	conn := registry.GetConnection(clientConn)
	assert.NotNil(t, conn)

	// Remove the connection
	registry.removeConnection(clientConn.ID())

	// Ensure the connection is removed
	removedConn := registry.GetConnection(clientConn)
	assert.NotEqual(t, conn, removedConn)
}
