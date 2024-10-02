package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
	"github.com/ksysoev/deriv-api-bff/pkg/middleware"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
)

func runServer(ctx context.Context, cfg *config) error {
	callhandler, err := handlers.NewCallHandler(&cfg.API)
	if err != nil {
		return fmt.Errorf("failed to create call handler: %w", err)
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

	server := api.NewSevice(&cfg.Server, requestHandler)

	return server.Run(ctx)
}
