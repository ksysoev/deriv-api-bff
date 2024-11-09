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
	"github.com/coder/websocket/wsjson"
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/core/handlerfactory"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	httpprov "github.com/ksysoev/deriv-api-bff/pkg/prov/http"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/router"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type testSuite struct {
	suite.Suite
	httpServ *httptest.Server
	mux      *http.ServeMux
}

// newTestSuite creates and returns a new instance of testSuite.
// It takes no parameters and returns a pointer to a testSuite instance.
func newTestSuite() *testSuite {
	return &testSuite{
		mux: http.NewServeMux(),
	}
}

// BeforeTest sets up the HTTP server and WebSocket echo server for the test suite.
// It takes two string parameters, which are not used in the function.
// It initializes the HTTP server and WebSocket echo server for testing purposes.
func (s *testSuite) BeforeTest(_, _ string) {
	s.mux = http.NewServeMux()
	s.mux.Handle("/ws", s.createTestWSEchoServer())
	s.httpServ = httptest.NewServer(s.mux)
}

// AfterTest performs cleanup operations after each test.
// It takes two unnamed string parameters which are not used in the function.
// It closes the HTTP server associated with the test suite to release resources.
func (s *testSuite) AfterTest(_, _ string) {
	s.httpServ.Close()
}

// echoWSURL returns the WebSocket URL for the echo server.
// It does not take any parameters.
// It returns a string which is the URL of the echo WebSocket server.
func (s *testSuite) echoWSURL() string {
	return s.httpServ.URL + "/ws"
}

func (s *testSuite) httpURL() string {
	return s.httpServ.URL
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
func (s *testSuite) startAppWithConfig(cfgYAML string) (url string, err error) {
	var cfg []handlerfactory.Config

	err = yaml.Unmarshal([]byte(cfgYAML), &cfg)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	handlers, err := createHandlers(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create handlers: %w", err)
	}

	derivAPI := deriv.NewService(&deriv.Config{Endpoint: s.echoWSURL()})
	connRegistry := repo.NewConnectionRegistry()
	calls := repo.NewCallsRepository()
	beRouter := router.New(derivAPI, httpprov.NewService())
	requestHandler := core.NewService(calls, beRouter, connRegistry)

	requestHandler.UpdateHandlers(handlers)

	server := api.NewSevice(&api.Config{Listen: ":0",
		RateLimits: api.RateLimits{
			General: api.GeneralRateLimits{
				Interval: "1s",
				Limit:    1000, // high rate limit for testing and benchmarking
			},
		},
	}, requestHandler)

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

		return "", fmt.Errorf("server did not start")
	}

	s.T().Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Millisecond):
		}
	})

	return fmt.Sprintf("ws://%s/?app_id=1", server.Addr().String()), nil
}

// testRequest sends a WebSocket request to the specified URL and asserts the response.
// It takes three parameters: url of type string, req of type any, and expectedResp of type any.
// It does not return any values but uses assertions to validate the response.
// It returns an error if the WebSocket connection fails, the request cannot be written, or the response cannot be read.
func (s *testSuite) testRequest(url string, req, expectedResp any) {
	a := assert.New(s.T())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	defer cancel()

	c, r, err := websocket.Dial(ctx, url, nil)
	a.NoError(err)

	if r.Body != nil {
		_ = r.Body.Close()
	}

	defer c.Close(websocket.StatusNormalClosure, "")

	err = wsjson.Write(ctx, c, &req)
	a.NoError(err)

	var resp map[string]any

	err = wsjson.Read(ctx, c, &resp)
	a.NoError(err)
	a.Equal(expectedResp, resp)
}

// DebugConfig marshals the provided configuration into YAML format and logs it.
// It takes cfg of type *cmd.Config.
// It does not return any values.
// It logs an error message and fails the test if marshalling the configuration fails.
func (s *testSuite) DebugConfig(cfg *cmd.Config) {
	d, err := yaml.Marshal(cfg)
	if err != nil {
		s.T().Fatalf("failed to marshal config: %v", err)
	}

	s.T().Logf("Config:\n%s", string(d))
}

func (s *testSuite) addHTTPContent(path, content string) {
	s.mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(content))
	})
}

func createHandlers(cfg []handlerfactory.Config) (map[string]core.Handler, error) {
	handlers := make(map[string]core.Handler, len(cfg))

	for _, c := range cfg {
		name, handler, err := handlerfactory.New(c)
		if err != nil {
			return nil, fmt.Errorf("failed to create handler: %w", err)
		}

		if _, ok := handlers[name]; ok {
			return nil, fmt.Errorf("duplicate handler name: %s", name)
		}

		handlers[name] = handler
	}

	return handlers, nil
}
