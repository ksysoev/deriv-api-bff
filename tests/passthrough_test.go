package tests

import (
	"context"
	"fmt"
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

func (s *testSuite) TestPassthrough() {
	a := assert.New(s.T())

	derivAPI := deriv.NewService(&deriv.Config{
		Endpoint: s.echoWSURL(),
	})

	connRegistry := repo.NewConnectionRegistry()

	calls, err := repo.NewCallsRepository(&repo.CallsConfig{})
	a.NoError(err)

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

		err := server.Run(ctx)
		a.NoError(err)

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

	c, r, err := websocket.Dial(ctx, wsURL, nil)
	a.NoError(err)

	if r.Body != nil {
		r.Body.Close()
	}

	defer c.Close(websocket.StatusNormalClosure, "")

	expectedResp := map[string]string{"ping": "1"}

	err = wsjson.Write(ctx, c, &expectedResp)
	a.NoError(err)

	var resp map[string]string
	err = wsjson.Read(ctx, c, &resp)
	a.NoError(err)
	a.Equal(expectedResp, resp)

	cancel()

	select {
	case <-done:
	case <-time.After(10 * time.Millisecond):
	}
}
