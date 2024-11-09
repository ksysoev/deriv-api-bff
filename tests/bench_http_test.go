package tests

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"golang.org/x/exp/slog"
)

func setup() *testSuite {
	t := new(testing.T)
	suite := newTestSuite()
	programLevel := new(slog.LevelVar)

	suite.SetT(t)
	suite.SetS(suite)
	suite.BeforeTest("", "")

	programLevel.Set(slog.LevelError)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	return suite
}

func Benchmark_HTTP_Params(b *testing.B) {
	suite := setup()

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
				"testcall": map[string]any{
					"data1": fmt.Sprintf("value%d", i),
					"data2": float64(i + 2),
				},
			}

			suite.testRequest(url, req, expectedResp)

			<-sem
			wg.Done()
		}()
	}

	wg.Wait()
}

func Benchmark_HTTP_Aggregation(b *testing.B) {
	suite := setup()

	defer suite.AfterTest("", "")

	httpURL := suite.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestAggregationConfig, "{{host}}", httpURL)

	url, err := suite.startAppWithConfig(cfg)
	if err != nil {
		suite.T().Fatal("failed to start app with config", err)
	}

	sem := make(chan struct{}, 5)
	wg := sync.WaitGroup{}

	b.ResetTimer()

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

	suite.addHTTPContent("/testcall1", `{"data1": "value1", "data2": 1}`)
	suite.addHTTPContent("POST /testcall2", `{"data1": "value2", "data2": 2}`)

	for i := 0; i < b.N; i++ {
		sem <- struct{}{}

		wg.Add(1)

		go func() {
			suite.testRequest(url, req, expectedResp)

			<-sem
			wg.Done()
		}()
	}

	wg.Wait()
}

func Benchmark_HTTP_Chain(b *testing.B) {
	suite := setup()

	defer suite.AfterTest("", "")

	httpURL := suite.httpURL()
	cfg := strings.ReplaceAll(testHTTPRequestChainConfig, "{{host}}", httpURL)

	url, err := suite.startAppWithConfig(cfg)
	if err != nil {
		suite.T().Fatal("failed to start app with config", err)
	}

	sem := make(chan struct{}, 5)
	wg := sync.WaitGroup{}

	b.ResetTimer()

	suite.addHTTPContent("/testcall1", `{"data1": "value1", "data2": 1}`)
	suite.addHTTPContent("POST /testcall2/value1", `{"data1": "value2", "data2": 2}`)

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

	for i := 0; i < b.N; i++ {
		sem <- struct{}{}

		wg.Add(1)

		go func() {
			suite.testRequest(url, req, expectedResp)

			<-sem
			wg.Done()
		}()
	}

	wg.Wait()
}
