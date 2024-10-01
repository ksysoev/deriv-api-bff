package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/deriv-api-bff/pkg/router"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/server"
)

const (
	Addr = ":8080"
)

func main() {

	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	config, err := handlers.LoadConfig("./config.yaml")
	if err != nil {
		slog.Error("Fail to load config", "error", err)
		os.Exit(1)
	}

	callhandler, err := handlers.NewCallHandler(config)

	if err != nil {
		slog.Error("Fail to create call handler", "error", err)
		os.Exit(1)
	}

	wsBackend := backend.NewWSBackend(
		"wss://ws.derivws.com/websockets/v3",
		func(r wasabi.Request) (wasabi.MessageType, []byte, error) {
			switch r.RoutingKey() {
			case "text":
				return wasabi.MsgTypeText, r.Data(), nil
			case "binary":
				return wasabi.MsgTypeBinary, r.Data(), nil
			default:
				var t wasabi.MessageType
				return t, nil, fmt.Errorf("unsupported request type: %s", r.RoutingKey())
			}
		},
		backend.WithWSDialler(func(ctx context.Context, baseURL string) (*websocket.Conn, error) {
			urlParams := middleware.QueryParamsFromContext(ctx)

			if urlParams != nil {
				if app_id := urlParams.Get("app_id"); app_id != "" {
					baseURL = fmt.Sprintf("%s?app_id=%s", baseURL, app_id)
				} else {
					return nil, fmt.Errorf("app_id is required")
				}

				if lang := urlParams.Get("l"); lang != "" {
					baseURL = fmt.Sprintf("%s&l=%s", baseURL, lang)
				}
			} else {
				return nil, fmt.Errorf("url params are required")
			}
			header := http.Header{}
			if h := middleware.HeadersFromContext(ctx); h != nil {
				header = h
			}

			c, resp, err := websocket.Dial(ctx, baseURL, &websocket.DialOptions{
				HTTPHeader: header,
			})

			if err != nil {
				return nil, err
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			return c, nil
		}),
	)

	requestHandler := handlers.NewBackendForFE(wsBackend, callhandler)

	dispatcher := dispatch.NewRouterDispatcher(requestHandler, router.Dispatch)
	endpoint := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))
	endpoint.Use(middleware.NewQueryParamsMiddleware())
	endpoint.Use(middleware.NewHeadersMiddleware())
	server := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	server.AddChannel(endpoint)

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}
