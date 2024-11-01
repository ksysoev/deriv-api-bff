package tests

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
          type: boolean
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
	url, err := s.startAppWithConfig(testRequestParamsConfig)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

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
		"param1":   "value1",
		"param2":   float64(2),
		"param3":   true,
	}

	s.testRequest(url, req, expectedResp)
}

const testAggergationConfig = `
server:
  listen: "localhost:0"
api:
  calls:
    - method: testcall
      backend:
        - response_body: data1
          request_template:
            data1:
              field1: value1
          allow: 
            - field1
        - response_body: data2
          request_template:
            data2:
              field2: value2
          allow:
            - field2
`

func (s *testSuite) TestAggregation() {
	url, err := s.startAppWithConfig(testAggergationConfig)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	req := map[string]any{"method": "testcall"}
	expectedResp := map[string]any{
		"echo":     req,
		"msg_type": "testcall",
		"field1":   "value1",
		"field2":   "value2",
	}

	s.testRequest(url, req, expectedResp)
}

const testChainConfig = `
server:
  listen: "localhost:0"
api:
  calls:
    - method: testcall
      backend:
        - response_body: data1
          request_template:
            data1:
              field1: value1
          allow: 
            - field1
        - response_body: data2
          depends_on:
            - data1
          request_template:
            data2:
              field2: ${resp.data1.field1}
          allow:
            - field2
`

func (s *testSuite) TestChain() {
	url, err := s.startAppWithConfig(testChainConfig)
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	req := map[string]any{"method": "testcall"}
	expectedResp := map[string]any{
		"echo":     req,
		"msg_type": "testcall",
		"field1":   "value1",
		"field2":   "value1",
	}

	s.testRequest(url, req, expectedResp)
}
