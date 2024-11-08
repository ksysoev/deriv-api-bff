//go:build !compile

package tests

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func Benchmark_HTTP_Chain(b *testing.B) {
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
			path1 := fmt.Sprintf("/testcall%d", i)
			path2 := fmt.Sprintf("POST /testcall%d/value%d", i+1, i)

			suite.addHTTPContent(path1, `{"data1": "value1", "data2": 1}`)
			suite.addHTTPContent(path2, `{"data1": "value2", "data2": 2}`)

			req := map[string]any{
				"method": "testcall",
			}
			expectedResp := map[string]any{
				"echo":     req,
				"msg_type": "testcall",
				"data1":    "value1",
				"data2":    float64(2),
			}

			suite.testRequest(url, req, expectedResp)

			<-sem
			wg.Done()
		}()
	}

	wg.Wait()
}
