package tests

import (
	"strings"
)

const testHTTRequestParamsConfig = `
- method: testcall
  params:
    param1:
        type: string
    param2:
        type: number
    param3:
        type: boolean
  backend:
    - name: testcall
      url: "{{host}}/testcall/${params.param1}/${params.param2}/${params.param3}"
      method: "GET"
      allow: 
        - data1
        - data2
`

func (s *testSuite) TestHTTPRequestParams() {
	httpURL := s.httpURL()
	cfg := strings.ReplaceAll(testHTTRequestParamsConfig, "{{host}}", httpURL)

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
		"testcall": map[string]any{
			"data1": "value1",
			"data2": float64(2),
		},
	}

	s.testRequest(url, req, expectedResp)
}

const testHTTPRequestAggregationConfig = `
- method: testcall
  backend:
    - name: testcall1
      url: "{{host}}/testcall1"
      method: GET
      allow: 
        - data1
    - name: testcall2
      url: "{{host}}/testcall2"
      method: POST
      allow: 
        - data2
`

func (s *testSuite) TestHTTPRequestAggregation() {
	httpURL := s.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestAggregationConfig, "{{host}}", httpURL)

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
		"testcall": map[string]any{
			"data1": "value1",
			"data2": float64(2),
		},
	}

	s.testRequest(url, req, expectedResp)
}

const testHTTPRequestChainConfig = `
- method: testcall
  backend:
    - name: testcall1
      url: "{{host}}/testcall1"
      method: GET
      allow: 
        - data1
    - name: testcall2
      depends_on:
        - testcall1
      url: "{{host}}/testcall2/${resp.testcall1.data1}"
      method: POST
      allow: 
        - data2
`

func (s *testSuite) TestHTTPRequestChain() {
	httpURL := s.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestChainConfig, "{{host}}", httpURL)

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
		"testcall": map[string]any{
			"data1": "value1",
			"data2": float64(2),
		},
	}

	s.testRequest(url, req, expectedResp)
}
