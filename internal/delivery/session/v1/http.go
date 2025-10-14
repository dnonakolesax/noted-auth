package session

import (
	"context"
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
)

type usecase interface {
	Get(ctx context.Context, token string) ([]byte, error)
	Delete(ctx context.Context, token string, id string) error
}

type Handler struct {
	sessionUsecase usecase
	logger         *slog.Logger
}

func NewSessionHandler(sesionUsecase usecase, logger *slog.Logger) *Handler {
	return &Handler{
		sessionUsecase: sesionUsecase,
		logger:         logger,
	}
}

// Get godoc
// @Summary Get all user sessions
// @Description Returns all user's sessions
// @Tags openid-connect injectable
// @Produces json
// @Success 200 {object} []model.Session
// @Failure 400
// @Failure 500
// @Router /session [get].
func (sh *Handler) Get(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	token := ctx.Request.Header.Cookie(consts.ATCookieKey)

	if token == nil {
		sh.logger.WarnContext(contex, "request sent without token")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	sessions, err := sh.sessionUsecase.Get(contex, string(token))

	if err != nil {
		sh.logger.ErrorContext(contex, "Error while getting sessions: ", "err", err.Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(sessions)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

// Delete godoc
// @Summary Delete session
// @Description Deletes session by id, or all session if id is not passed
// @Tags openid-connect injectable
// @Param id path string false "Session ID"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /session/{id} [delete].
func (sh *Handler) Delete(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	token := ctx.Request.Header.Cookie(consts.ATCookieKey)

	if token == nil {
		sh.logger.WarnContext(contex, "request sent without token")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	sessionID := ctx.UserValue("id")

	sidStr := ""
	if sessionID != nil {
		var ok bool
		sidStr, ok = sessionID.(string)
		if !ok {
			sh.logger.ErrorContext(contex, "Error while casting sessionId to string", slog.Any("sessionID", sessionID))
		}
	}

	err := sh.sessionUsecase.Delete(contex, string(token), sidStr)

	if err != nil {
		sh.logger.ErrorContext(contex, "Error while deleting session", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (sh *Handler) RegisterRoutes(apiGroup *router.Group) {
	g := apiGroup.Group("/session")
	g.GET("", sh.Get)
	g.DELETE("/{id}", sh.Delete)
}
