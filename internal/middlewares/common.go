package middlewares

import (
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/cryptos"
)

const requestIDSize = 16

func CommonMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		requestID := ctx.Request.Header.Peek("X-Request-Id")
		var reqID string
		if requestID == nil {
			var err error
			requestID, err = cryptos.GenRandomString(requestIDSize)
			reqID = base64.RawURLEncoding.EncodeToString(requestID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error generating RequestId: %v", err))
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				return
			}
			ctx.Request.Header.Set("X-Request-Id", reqID)
		} else {
			reqID = string(requestID)
		}
		slog.Info("Received Request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("ip", ctx.RemoteIP().String()),
			slog.String("requestId", reqID),
			slog.String("userAgent", string(ctx.UserAgent())),
		)
		now := ctx.Time().UnixMilli()
		h(ctx)
		end := ctx.Time().UnixMilli()
		slog.Info("Completed request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("requestId", reqID),
			slog.Int("status", ctx.Response.StatusCode()),
			slog.Int("duration", int(end-now)),
		)
	})
}
