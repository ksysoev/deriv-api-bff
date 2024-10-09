package api

import (
	"context"
	"testing"
	"time"
)

func TestSvc_Run(t *testing.T) {
	config := &Config{
		Listen: ":8080",
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

// func TestSvc_MultipleRun(t *testing.T) {
// 	mockRequestHandler := wasabi_mocks.NewMockRequestHandler(t)
// 	config := &Config{
// 		Listen: ":8081",
// 	}

// 	service := NewSevice(config, mockRequestHandler)
// 	ctx, cancel := context.WithCancel(context.Background())

// 	done := make(chan struct{})

// 	go func() {
// 		err := service.Run(ctx)
// 		switch err {
// 		case nil:
// 			close(done)
// 		default:
// 			t.Errorf("got unexpected error: %s", err)
// 		}
// 	}()

// 	go func() {
// 		cancel()
// 	}()

// 	select {
// 	case <-done:
// 	case <-time.After(1 * time.Second):
// 		t.Error("Expected server to stop")
// 	}
// }
