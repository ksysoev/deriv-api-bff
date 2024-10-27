package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/core"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	httpprov "github.com/ksysoev/deriv-api-bff/pkg/prov/http"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/router"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/stretchr/testify/assert"
)

func createTestWSEchoServer(_ *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ws echo server")
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

			_, err = io.Copy(wsw, wsr)
			if err != nil {
				assert.NoError(nil, err)
				return
			}

			err = wsw.Close()
			if err != nil {
				assert.NoError(nil, err)
				return
			}
		}
	})
}

func TestPassthrough(t *testing.T) {
	ts := httptest.NewServer(createTestWSEchoServer(t))
	defer ts.Close()

	derivAPI := deriv.NewService(&deriv.Config{
		Endpoint: ts.URL,
	})

	connRegistry := repo.NewConnectionRegistry()

	calls, err := repo.NewCallsRepository(&repo.CallsConfig{})
	assert.NoError(t, err)

	beRouter := router.New(derivAPI, httpprov.NewService())
	requestHandler := core.NewService(calls, beRouter, connRegistry)

	server := api.NewSevice(&api.Config{
		Listen: "localhost:0",
	}, requestHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})
	done := make(chan struct{})

	go func() {
		close(ready)
		server.Run(ctx)
		close(done)
	}()

	select {
	case <-ready:
	case <-time.After(10 * time.Millisecond):
	}

	for server.Addr() == nil {
		time.Sleep(10 * time.Millisecond)
	}

	wsURL := fmt.Sprintf("ws://%s/?app_id=1", server.Addr())

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	assert.NoError(t, err)

	defer c.Close(websocket.StatusNormalClosure, "")

	expectedResp := map[string]string{"ping": "1"}

	err = wsjson.Write(ctx, c, &expectedResp)
	assert.NoError(t, err)

	var resp map[string]string
	err = wsjson.Read(ctx, c, &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)

	cancel()

	select {
	case <-done:
	case <-time.After(10 * time.Millisecond):
	}
}
