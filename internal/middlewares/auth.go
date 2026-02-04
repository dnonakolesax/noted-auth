package middlewares

import (
	"context"
	"log/slog"

	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/cookies"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type IntrospectUsecase interface {
	GetUserID(ctx context.Context, at string, rt string) (model.TokenGRPCDTO, error)
}

type AuthMW struct {
	usecase IntrospectUsecase
	logger  *slog.Logger
}

func NewAuthMW(usecase IntrospectUsecase, logger *slog.Logger) *AuthMW {
	return &AuthMW{usecase: usecase, logger: logger}
}

func (am *AuthMW) AuthMiddleware(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
		contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
		at := ctx.Request.Header.Cookie(consts.ATCookieKey)
		if at == nil {
			am.logger.WarnContext(contex, "no at passed")
		}
		rt := ctx.Request.Header.Cookie(consts.RTCookieKey)
		if rt == nil {
			am.logger.WarnContext(contex, "no rt passed")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}
		dto, err := am.usecase.GetUserID(context.Background(), string(at), string(rt))

		if err != nil {
			am.logger.ErrorContext(contex, "error introspecting", slog.String(consts.ErrorLoggerKey, err.Error()))
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}
		ctx.Request.SetUserValue(consts.CtxUserIDKey, dto.UserID)
		if dto.AccessToken != "" && dto.RefreshToken != "" && dto.IDToken != "" {
			cookies.SetupAccessCookies(ctx, dto.ToTokenDTO())
		}
		am.logger.Debug(dto.AccessToken)
		am.logger.Debug(dto.RefreshToken)
		am.logger.Debug(dto.IDToken)
		h(ctx)
	})
}
