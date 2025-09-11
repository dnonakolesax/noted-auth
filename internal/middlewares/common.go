package middlewares

import (
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/cryptos"
	"github.com/valyala/fasthttp"
)

func CommonMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		requestId := ctx.Request.Header.Peek("X-Request-Id")
		var reqId string
		if requestId == nil {
			var err error
			requestId, err = cryptos.GenRandomString(16)
			reqId = base64.RawURLEncoding.EncodeToString(requestId)
			if err != nil {
				slog.Error(fmt.Sprintf("Error generating RequestId: %v", err))
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				return
			}
			ctx.Request.Header.Set("X-Request-Id", reqId)
		} else {
			reqId = string(requestId)
		}
		slog.Info("Received Request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("ip", ctx.RemoteIP().String()),
			slog.String("requestId", reqId),
			slog.String("userAgent", string(ctx.UserAgent())),
		)
		now := ctx.Time().UnixMilli()
		h(ctx)
		end := ctx.Time().UnixMilli()
		slog.Info("Completed request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("requestId", reqId),
			slog.Int("status", ctx.Response.StatusCode()),
			slog.Int("duration", int(end-now)),
		)
	})
}
