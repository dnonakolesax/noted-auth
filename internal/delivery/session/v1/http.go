package session

import (
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
)

type usecase interface {
	Get(token string) ([]byte, error)
	Delete(token string, id string) error
}

type Handler struct {
	SessionUsecase usecase
}

func NewSessionHandler(sesionUsecase usecase) *Handler {
	return &Handler{
		SessionUsecase: sesionUsecase,
	}
}

// Get godoc
// @Summary Get all user sessions
// @Description Returns all user's sessions
// @Tags openid-connect
// @Produces json
// @Success 200 {object} []model.Session
// @Failure 400
// @Failure 500
// @Router /session [get].
func (sh *Handler) Get(ctx *fasthttp.RequestCtx) {
	token := ctx.Request.Header.Cookie(consts.ATCookieKey)

	if token == nil {
		slog.Warn("request sent without token")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	sessions, err := sh.SessionUsecase.Get(string(token))

	if err != nil {
		slog.Error("Error while getting sessions: ", "err", err.Error())
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
// @Tags openid-connect
// @Param id path string false "Session ID"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /session/{id} [delete].
func (sh *Handler) Delete(ctx *fasthttp.RequestCtx) {
	token := ctx.Request.Header.Cookie(consts.ATCookieKey)

	if token == nil {
		slog.Warn("request sent without token")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	sessionID := ctx.UserValue("id")

	sidStr := ""
	if sessionID != nil {
		var ok bool
		sidStr, ok = sessionID.(string)
		if !ok {
			slog.Error("Error while casting sessionId to string", "err", sidStr)
		}
	}

	err := sh.SessionUsecase.Delete(string(token), sidStr)

	if err != nil {
		slog.Error("Error while deleting session", "err", err.Error())
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
