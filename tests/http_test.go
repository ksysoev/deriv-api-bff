package tests

import (
	"strings"
)

const testHTTRequestParamsConfig = `
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
        - name: testcall
          url_template: "{{host}}/testcall/${params.param1}/${params.param2}/${params.param3}"
          method: "GET"
          allow: 
            - data1
            - data2
`

func (s *testSuite) TestHTTPRequestParams() {
	httpUrl := s.httpURL()
	cfg := strings.ReplaceAll(testHTTRequestParamsConfig, "{{host}}", httpUrl)

	url, err := s.startAppWithConfig(cfg)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	s.addHTTPContent("GET /testcall/value1/2/true", `{"data1": "value1", "data2": 2}`)

	req := map[string]any{
		"method": "testcall",
		"params": map[string]any{
			"param1": "value1",
			"param2": float64(2),
			"param3": true,
		},
	}
	expectedResp := map[string]any{
		"echo":     req,
		"msg_type": "testcall",
		"data1":    "value1",
		"data2":    float64(2),
	}

	s.testRequest(url, req, expectedResp)
}

const testHTTPRequestAggregationConfig = `
server:
  listen: "localhost:0"
api:
  calls:
    - method: testcall
      backend:
        - name: testcall1
          url_template: "{{host}}/testcall1"
          method: GET
          allow: 
            - data1
        - name: testcall2
          url_template: "{{host}}/testcall2"
          method: POST
          allow: 
            - data2
`

func (s *testSuite) TestHTTPRequestAggregation() {
	httpUrl := s.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestAggregationConfig, "{{host}}", httpUrl)

	url, err := s.startAppWithConfig(cfg)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	s.addHTTPContent("/testcall1", `{"data1": "value1", "data2": 1}`)
	s.addHTTPContent("POST /testcall2", `{"data1": "value2", "data2": 2}`)

	req := map[string]any{
		"method": "testcall",
	}
	expectedResp := map[string]any{
		"echo":     req,
		"msg_type": "testcall",
		"data1":    "value1",
		"data2":    float64(2),
	}

	s.testRequest(url, req, expectedResp)
}

const testHTTPRequestChainConfig = `
server:
  listen: "localhost:0"
api:
  calls:
    - method: testcall
      backend:
        - name: testcall1
          url_template: "{{host}}/testcall1"
          method: GET
          allow: 
            - data1
        - name: testcall2
          depends_on:
            - testcall1
          url_template: "{{host}}/testcall2/${resp.testcall1.data1}"
          method: POST
          allow: 
            - data2
`

func (s *testSuite) TestHTTPRequestChain() {
	httpUrl := s.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestChainConfig, "{{host}}", httpUrl)

	url, err := s.startAppWithConfig(cfg)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	s.addHTTPContent("/testcall1", `{"data1": "value1", "data2": 1}`)
	s.addHTTPContent("POST /testcall2/value1", `{"data1": "value2", "data2": 2}`)

	req := map[string]any{
		"method": "testcall",
	}
	expectedResp := map[string]any{
		"echo":     req,
		"msg_type": "testcall",
		"data1":    "value1",
		"data2":    float64(2),
	}

	s.testRequest(url, req, expectedResp)
}
