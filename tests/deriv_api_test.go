package tests

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ksysoev/deriv-api-bff/pkg/cmd"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const testRequestParamsConfig = `
server:
  listen: "localhost:0"
api:
  calls:
    - method: testcall
      params:
        param1:
          type: string
        param2:
          type: number
        param3:
          type: bool
      backend:
        - response_body: "data"
          request_template:
            data:
              param1: "${params.param1}"
              param2: "${params.param2}"
              param3: "${params.param3}"
          allow: 
            - param1
            - param2
            - param3
`

func (s *testSuite) TestRequestParams() {
	a := assert.New(s.T())

	var cfg cmd.Config

	err := yaml.Unmarshal([]byte(testRequestParamsConfig), &cfg)
	a.NoError(err)

	cfg.Deriv.Endpoint = s.echoWSURL()

	url, stopServer, err := s.startAppWithConfig(&cfg)

	a.NoError(err)

	defer stopServer()

	ctx := context.Background()

	c, r, err := websocket.Dial(ctx, url, nil)
	a.NoError(err)

	if r.Body != nil {
		r.Body.Close()
	}

	defer c.Close(websocket.StatusNormalClosure, "")

	expectedResp := map[string]any{
		"method": "testcall",
		"params": map[string]any{
			"param1": "value1",
			"param2": float64(2),
			"param3": true,
		},
	}

	err = wsjson.Write(ctx, c, &expectedResp)
	a.NoError(err)

	var resp map[string]any
	err = wsjson.Read(ctx, c, &resp)
	a.NoError(err)

	a.Equal(
		map[string]any{
			"echo":     expectedResp,
			"msg_type": "testcall",
			"param1":   "value1",
			"param2":   float64(2),
			"param3":   true,
		},
		resp,
	)
}
