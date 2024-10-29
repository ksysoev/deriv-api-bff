package tests

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
	"github.com/stretchr/testify/assert"
)

func (s *testSuite) TestPassthrough() {
	a := assert.New(s.T())

	url, stopServer, err := s.startAppWithConfig(&cmd.Config{
		Server: api.Config{
			Listen: "localhost:0",
		},
		API: repo.CallsConfig{},
		Deriv: deriv.Config{
			Endpoint: s.echoWSURL(),
		},
	})

	a.NoError(err)

	defer stopServer()

	ctx := context.Background()

	c, r, err := websocket.Dial(ctx, url, nil)
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
}
