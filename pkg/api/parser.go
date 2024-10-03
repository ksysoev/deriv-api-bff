package api

import (
	"context"
	"encoding/json"

	"github.com/ksysoev/deriv-api-bff/pkg/router"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

func parser(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
	if msgType == wasabi.MsgTypeBinary {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	var req router.Request

	if err := json.Unmarshal(data, &req); err != nil {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	if req.Method == "" {
		return dispatch.NewRawRequest(ctx, msgType, data)
	}

	req.Ctx = ctx
	req.RawData = data

	return &req
}
