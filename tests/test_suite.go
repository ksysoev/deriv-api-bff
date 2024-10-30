//go:build !compile

package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/config"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	httpprov "github.com/ksysoev/deriv-api-bff/pkg/prov/http"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/router"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	echoWS *httptest.Server
}

// newTestSuite creates and returns a new instance of testSuite.
// It takes no parameters and returns a pointer to a testSuite instance.
func newTestSuite() *testSuite {
	return &testSuite{}
}

// SetupSuite initializes the test suite by setting up a new WebSocket server.
// It does not take any parameters and does not return any values.
// This function is typically called before any tests are run to ensure the test environment is properly configured.
func (s *testSuite) SetupSuite() {
	s.echoWS = httptest.NewServer(s.createTestWSEchoServer())
}

// TearDownSuite closes the WebSocket connection used by the test suite.
// It does not take any parameters and does not return any values.
// This function ensures that the WebSocket connection is properly closed after all tests in the suite have run.
func (s *testSuite) TearDownSuite() {
	s.echoWS.Close()
}

// echoWSURL returns the WebSocket URL for the echo server.
// It does not take any parameters.
// It returns a string which is the URL of the echo WebSocket server.
func (s *testSuite) echoWSURL() string {
	return s.echoWS.URL
}

// createTestWSEchoServer creates a WebSocket echo server handler function.
// It returns an http.HandlerFunc that establishes a WebSocket connection,
// reads messages from the client, and echoes them back.
// If an error occurs during the WebSocket handshake or message processing,
// the connection is closed gracefully.
func (s *testSuite) createTestWSEchoServer() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		for {
			_, wsr, err := c.Reader(r.Context())
			if err != nil {
				if err == io.EOF {
					return
				}

				return
			}

			wsw, err := c.Writer(r.Context(), websocket.MessageText)
			if err != nil {
				return
			}

			if _, err := io.Copy(wsw, wsr); err != nil {
				return
			}

			if err := wsw.Close(); err != nil {
				return
			}
		}
	})
}

// startAppWithConfig starts the application with the provided configuration.
// It takes cfg of type *cmd.Config which contains the configuration settings for the application.
// It returns a string representing the URL of the started server, a function to close the server, and an error if the server fails to start.
// It returns an error if the calls repository cannot be created or if the server does not start within the specified timeout.
func (s *testSuite) startAppWithConfig(cfg *config.Config) (url string, closer func(), err error) {
	derivAPI := deriv.NewService(&cfg.Deriv)

	connRegistry := repo.NewConnectionRegistry()
	event := config.NewEvent[config.CallsConfig]()

	calls, err := repo.NewCallsRepository(&cfg.API, event)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create calls repo: %w", err)
	}

	beRouter := router.New(derivAPI, httpprov.NewService())
	requestHandler := core.NewService(calls, beRouter, connRegistry)

	server := api.NewSevice(&cfg.Server, requestHandler)

	ctx, cancel := context.WithCancel(context.Background())

	ready := make(chan struct{})
	done := make(chan struct{})

	go func() {
		for server.Addr() == nil {
			time.Sleep(10 * time.Millisecond)
		}

		close(ready)
	}()

	go func() {
		err := server.Run(ctx)
		assert.NoError(s.T(), err)

		close(done)
	}()

	select {
	case <-ready:
	case <-time.After(time.Second):
		cancel()

		return "", nil, fmt.Errorf("server did not start")
	}

	closer = func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Millisecond):
		}
	}

	url = fmt.Sprintf("ws://%s/?app_id=1", server.Addr().String())

	return url, closer, nil
}
