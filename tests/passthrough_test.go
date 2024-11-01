package tests

func (s *testSuite) TestPassthrough() {
	url, err := s.startAppWithConfig("server:\n  listen: localhost:0\n")
	if err != nil {
		s.T().Fatal("failed to start app with config", err)
	}

	req := map[string]any{"ping": "1"}
	expectedResp := req

	s.testRequest(url, req, expectedResp)
}
