package main

import (
	"context"
	"fmt"
	"log/slog"
	_ "net/http/pprof"
	"os"

	"github.com/ksysoev/deriv-api-bff/pkg/handlers"
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

	backend := backend.NewWSBackend(
		"wss://ws.derivws.com/websockets/v3?app_id=1089",
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
	)

	requestHandler := handlers.NewBackendForFE(backend, callhandler)

	dispatcher := dispatch.NewRouterDispatcher(requestHandler, router.Dispatch)
	channel := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	server := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	server.AddChannel(channel)

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}
