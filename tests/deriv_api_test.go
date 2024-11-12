package tests

const testRequestParamsConfig = `
- method: testcall
  params:
    param1:
      type: string
    param2:
      type: number
    param3:
      type: boolean
  backend:
    - request: 
        data:
            param1: ${params.param1}
            param2: ${params.param2}
            param3: ${params.param3}
        msg_type: data
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
		"testcall": map[string]any{
			"param1": "value1",
			"param2": float64(2),
			"param3": true,
		},
	}

	s.testRequest(url, req, expectedResp)
}

const testAggergationConfig = `
- method: testcall
  backend:
    - request:
        data1:
            field1: value1
        msg_type: data1
      allow: 
        - field1
    - request:
        data2:
            field2: value2
        msg_type: data2
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
		"testcall": map[string]any{
			"field1": "value1",
			"field2": "value2",
		},
	}

	s.testRequest(url, req, expectedResp)
}

const testChainConfig = `
- method: testcall
  backend:
    - name: data1
      request:
        data1:
            field1: value1
        msg_type: data1
      allow: 
        - field1
    - depends_on:
        - data1
      request:
        data2:
            field2: ${resp.data1.field1}
        msg_type: data2
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
		"testcall": map[string]any{
			"field1": "value1",
			"field2": "value1",
		},
	}

	s.testRequest(url, req, expectedResp)
}
