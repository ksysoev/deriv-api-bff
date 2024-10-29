//go:build !compile

package tests

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/coder/websocket"
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

				assert.NoError(nil, err)

				return
			}

			wsw, err := c.Writer(r.Context(), websocket.MessageText)
			if err != nil {
				assert.NoError(nil, err)

				return
			}

			if _, err := io.Copy(wsw, wsr); err != nil {
				assert.NoError(nil, err)

				return
			}

			if err := wsw.Close(); err != nil {
				assert.NoError(nil, err)

				return
			}
		}
	})
}
