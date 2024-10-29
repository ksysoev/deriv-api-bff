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
	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
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

func newTestSuite() *testSuite {
	return &testSuite{}
}

func (s *testSuite) SetupSuite() {
	s.echoWS = httptest.NewServer(s.createTestWSEchoServer())
}

func (s *testSuite) TearDownSuite() {
	s.echoWS.Close()
}

func (s *testSuite) echoWSURL() string {
	return s.echoWS.URL
}

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

func (s *testSuite) startAppWithConfig(cfg *cmd.Config) (url string, closer func(), err error) {
	derivAPI := deriv.NewService(&cfg.Deriv)

	connRegistry := repo.NewConnectionRegistry()

	calls, err := repo.NewCallsRepository(&cfg.API)
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
