package tests

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func Benchmark_HTTP_Params(b *testing.B) {
	t := new(testing.T)
	suite := newTestSuite()

	suite.SetT(t)
	suite.SetS(suite)

	suite.BeforeTest("", "")

	defer suite.AfterTest("", "")

	httpURL := suite.httpURL()
	cfg := strings.ReplaceAll(testHTTRequestParamsConfig, "{{host}}", httpURL)
	url, err := suite.startAppWithConfig(cfg)

	if err != nil {
		b.Errorf("Error starting server: %v", err)
	}

	sem := make(chan struct{}, 5)
	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sem <- struct{}{}

		wg.Add(1)

		go func() {
			suite.addHTTPContent(fmt.Sprintf("GET /testcall/value%d/%d/true", i, i+1), `{"data1": "value1", "data2": 2}`)

			req := map[string]any{
				"method": "testcall",
				"params": map[string]any{
					"param1": fmt.Sprintf("value%d", i),
					"param2": float64(i + 1),
					"param3": true,
				},
			}
			expectedResp := map[string]any{
				"echo":     req,
				"msg_type": "testcall",
				"data1":    fmt.Sprintf("value%d", i),
				"data2":    float64(i + 2),
			}

			suite.testRequest(url, req, expectedResp)

			<-sem
			wg.Done()
		}()
	}

	wg.Wait()
}
