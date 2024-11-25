package api

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/deriv-api-bff/pkg/core/request"
	wasabi "github.com/ksysoev/wasabi"
	httpmid "github.com/ksysoev/wasabi/middleware/http"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewSevice(t *testing.T) {
	cfg := &Config{
		Listen: "localhost:8080",
	}

	mockBFFService := NewMockBFFService(t)

	svc, err := NewSevice(cfg, mockBFFService)

	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, cfg, svc.cfg)
	assert.Equal(t, mockBFFService, svc.handler)
}

func TestSvc_Run(t *testing.T) {
	config := &Config{
		Listen: ":0",
	}

	service, err := NewSevice(config, nil)
	ctx, cancel := context.WithCancel(context.Background())

	assert.NoError(t, err)
	cancel()

	defer cancel()

	done := make(chan struct{})

	// Run the server
	go func() {
		err := service.Run(ctx)
		switch err {
		case nil:
			close(done)
		default:
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to stop")
	}
}

func TestSvc_Error(t *testing.T) {
	config := &Config{
		Listen: ":0",
		RateLimits: RateLimits{
			General: GeneralRateLimits{
				Interval: "invalid",
			},
		},
	}

	service, err := NewSevice(config, nil)
	assert.Error(t, err)
	assert.Nil(t, service)
}

func TestParse(t *testing.T) {
	tests := []struct {
		expected wasabi.Request
		name     string
		data     []byte
		msgType  wasabi.MessageType
	}{
		{
			name:     "TextMessage",
			msgType:  wasabi.MsgTypeText,
			data:     []byte("test text message"),
			expected: request.NewRequest(context.Background(), request.TextMessage, []byte("test text message")),
		},
		{
			name:     "BinaryMessage",
			msgType:  wasabi.MsgTypeBinary,
			data:     []byte{0x01, 0x02, 0x03},
			expected: request.NewRequest(context.Background(), request.BinaryMessage, []byte{0x01, 0x02, 0x03}),
		},
		{
			name:     "UnsupportedMessageType",
			msgType:  wasabi.MessageType(999),
			data:     []byte("unsupported message"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := parse(nil, context.Background(), tt.msgType, tt.data)
			assert.Equal(t, tt.expected, req)
		})
	}
}

func TestService_Addr(t *testing.T) {
	cfg := &Config{
		Listen: "localhost:0",
	}

	mockBFFService := NewMockBFFService(t)
	service, err := NewSevice(cfg, mockBFFService)

	assert.NoError(t, err)

	addr := service.Addr()
	assert.Nil(t, addr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan struct{})

	go func() {
		for service.Addr() == nil {
			time.Sleep(10 * time.Millisecond)
		}

		close(ready)
	}()

	go func() {
		_ = service.Run(ctx)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to start")
	}

	addr = service.Addr()
	assert.NotNil(t, addr)
}
func TestPopulateDefaults(t *testing.T) {
	tests := []struct {
		input    *Config
		expected *Config
		name     string
	}{
		{
			name: "Defaults Applied",
			input: &Config{
				Listen: "localhost:8080",
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        maxRequestsDefault,
				MaxRequestsPerConn: maxRequestsPerConnDefault,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: generalRateLimitIntervalDefault,
						Limit:    generalRateLimitDefault,
					},
				},
			},
		},
		{
			name: "MaxRequests Set",
			input: &Config{
				Listen:      "localhost:8080",
				MaxRequests: 200,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: maxRequestsPerConnDefault,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: generalRateLimitIntervalDefault,
						Limit:    generalRateLimitDefault,
					},
				},
			},
		},
		{
			name: "MaxRequestsPerConn Set",
			input: &Config{
				Listen:             "localhost:8080",
				MaxRequestsPerConn: 20,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        maxRequestsDefault,
				MaxRequestsPerConn: 20,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: generalRateLimitIntervalDefault,
						Limit:    generalRateLimitDefault,
					},
				},
			},
		},
		{
			name: "General Rate Limit Set",
			input: &Config{
				Listen:             "localhost:8080",
				MaxRequestsPerConn: 20,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: "1h",
						Limit:    1000,
					},
				},
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        maxRequestsDefault,
				MaxRequestsPerConn: 20,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: "1h",
						Limit:    1000,
					},
				},
			},
		},
		{
			name: "All Values Set",
			input: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: 20,
			},
			expected: &Config{
				Listen:             "localhost:8080",
				MaxRequests:        200,
				MaxRequestsPerConn: 20,
				RateLimits: RateLimits{
					General: GeneralRateLimits{
						Interval: generalRateLimitIntervalDefault,
						Limit:    generalRateLimitDefault,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			populateDefaults(tt.input)
			assert.Equalf(t, tt.expected, tt.input, tt.name)
		})
	}
}

func TestSkipMetrics(t *testing.T) {
	tests := []struct {
		name       string
		routingKey string
		want       bool
	}{
		{
			name:       "TextMessage routing key",
			routingKey: request.TextMessage,
			want:       true,
		},
		{
			name:       "BinaryMessage routing key",
			routingKey: request.BinaryMessage,
			want:       true,
		},
		{
			name:       "Other routing key",
			routingKey: "other",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := mocks.NewMockRequest(t)
			mockRequest.EXPECT().RoutingKey().Return(tt.routingKey)

			got := skipMetrics(mockRequest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRequestLimits_Success(t *testing.T) {
	validRateLimits := RateLimits{
		General: GeneralRateLimits{
			Interval: "10s",
			Limit:    200,
		},
	}
	requestLimitFunc, err := getRequestLimits(validRateLimits)
	mockRequest := mocks.NewMockRequest(t)
	ctx := context.Background()

	mockRequest.EXPECT().Context().Return(ctx)

	key, duration, limit := requestLimitFunc(mockRequest)

	assert.NoError(t, err)
	assert.Equal(t, "nil", key)
	assert.Equal(t, 10*time.Second, duration)
	assert.Equal(t, uint64(200), limit)
}

func TestGetRequestLimits_Failure(t *testing.T) {
	validRateLimits := RateLimits{
		General: GeneralRateLimits{
			Interval: "invalid",
		},
	}
	requestLimitFunc, err := getRequestLimits(validRateLimits)

	assert.Error(t, err)
	assert.Nil(t, requestLimitFunc)
}

func TestGetIPFromRequest_OK(t *testing.T) {
	tests := []struct {
		inputCtx   context.Context
		expectedIP string
	}{
		{
			inputCtx:   context.WithValue(context.Background(), httpmid.ClientIP, "8.8.8.8"),
			expectedIP: "8.8.8.8",
		},
		{
			inputCtx:   context.Background(),
			expectedIP: "nil",
		},
	}

	for _, test := range tests {
		mockRequest := mocks.NewMockRequest(t)
		mockRequest.EXPECT().Context().Return(test.inputCtx)
		ip := getIPFromRequest(mockRequest)
		assert.Equal(t, test.expectedIP, ip)
	}
}

var testGroups = []GroupRateLimits{
	{
		Name: "group1",
		Limits: GeneralRateLimits{
			Interval: "10s",
			Limit:    1000,
		},
		Methods: []string{"aggregate"},
	},
	{
		Name: "group2",
		Limits: GeneralRateLimits{
			Interval: "3m",
			Limit:    10,
		},
		Methods: []string{"chain"},
	},
}

func Test_buildGroupRateMap(t *testing.T) {
	groupRatesMap, err := buildGroupRateMap(testGroups)

	assert.NoError(t, err)
	assert.Equal(t, groupRatesMap, map[string]GroupRateLimits{
		"aggregate": {Name: "group1", Methods: []string{"aggregate"}, Limits: GeneralRateLimits{Interval: "10s", Limit: 1000}},
		"chain":     {Name: "group2", Methods: []string{"chain"}, Limits: GeneralRateLimits{Interval: "3m", Limit: 10}},
	})
}

func Test_buildGroupRateMap_Err(t *testing.T) {
	groupRatesMap, err := buildGroupRateMap(append(testGroups, GroupRateLimits{
		Name:    "group3",
		Methods: []string{"chain"},
	}))

	assert.Error(t, err)
	assert.Nil(t, groupRatesMap)
}

func Test_getRateLimitForMethods(t *testing.T) {
	requestLimitFunc, err := getRequestLimits(RateLimits{Groups: testGroups,
		General: GeneralRateLimits{Interval: "1s", Limit: 100}})
	mockRequest := mocks.NewMockRequest(t)
	ctx := context.WithValue(context.Background(), httpmid.ClientIP, "8.8.8.8")

	mockRequest.EXPECT().Context().Return(ctx)
	mockRequest.EXPECT().RoutingKey().Return("aggregate")
	assert.NoError(t, err)

	key, duration, limit := requestLimitFunc(mockRequest)

	assert.Equal(t, "8.8.8.8", key)
	assert.Equal(t, 10*time.Second, duration)
	assert.Equal(t, uint64(1000), limit)
}

func Test_getRateLimitForMethods_FromGeneral(t *testing.T) {
	requestLimitFunc, err := getRequestLimits(RateLimits{Groups: testGroups,
		General: GeneralRateLimits{Interval: "1s", Limit: 1000}})
	mockRequest := mocks.NewMockRequest(t)
	ctx := context.Background()

	mockRequest.EXPECT().Context().Return(ctx)
	mockRequest.EXPECT().RoutingKey().Return("undefined")
	assert.NoError(t, err)

	key, duration, limit := requestLimitFunc(mockRequest)

	assert.Equal(t, "nil", key)
	assert.Equal(t, 1*time.Second, duration)
	assert.Equal(t, uint64(1000), limit)
}

func Test_getRateLimitForMethods_Default(t *testing.T) {
	requestLimitFunc, err := getRequestLimits(RateLimits{Groups: append(testGroups,
		GroupRateLimits{
			Name:    "group3",
			Methods: []string{"config"},
		}),
		General: GeneralRateLimits{Interval: "1s", Limit: 1000}})
	mockRequest := mocks.NewMockRequest(t)
	ctx := context.Background()

	mockRequest.EXPECT().Context().Return(ctx)
	mockRequest.EXPECT().RoutingKey().Return("config")
	assert.NoError(t, err)

	key, duration, limit := requestLimitFunc(mockRequest)

	assert.Equal(t, "nil", key)
	assert.Equal(t, 1*time.Millisecond, duration)
	assert.Equal(t, uint64(100000), limit)
}
