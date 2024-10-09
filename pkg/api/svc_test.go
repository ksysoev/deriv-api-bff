package api

import (
	"context"
	"testing"
	"time"
)

func TestSvc_Run(t *testing.T) {
	config := &Config{
		Listen: ":0",
	}

	service := NewSevice(config, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	defer cancel()

	done := make(chan struct{})

	// Run the server
	go func() {
		err := service.Run(ctx)
		switch err {
		case nil:
			close(done)
		default:
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("Expected server to stop")
	}
}
