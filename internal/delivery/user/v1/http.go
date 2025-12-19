package user

import (
	"context"
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/cookies"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	Get(ctx context.Context, uuid string) (model.User, error)
}

type aUsecase interface {
	GetUserID(ctx context.Context, at string, rt string) (model.TokenGRPCDTO, error)
}

type Handler struct {
	userUsecase usecase
	authUsecase aUsecase
	logger      *slog.Logger
}

func NewUserHandler(userUsecase usecase, logger *slog.Logger, authUsecase aUsecase) *Handler {
	return &Handler{
		userUsecase: userUsecase,
		logger:      logger,
		authUsecase: authUsecase,
	}
}

// Get godoc
// @Summary Get user info
// @Description Returns user's name, surname and username
// @Tags openid-connect
// @Param id path string true "User ID"
// @Produces json
// @Success 200 {object} model.User
// @Failure 400
// @Failure 500
// @Router /users/{id} [get].
func (uh *Handler) Get(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	userID := ctx.UserValue("id")

	if userID == nil {
		uh.logger.WarnContext(contex, "empty user id")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	idString, ok := userID.(string)

	if !ok {
		uh.logger.ErrorContext(contex, "could not convert user id to string")
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	user, err := uh.userUsecase.Get(contex, idString)

	if err != nil {
		uh.logger.WarnContext(contex, "could not get user", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userJSON, err := user.MarshalJSON()

	if err != nil {
		uh.logger.ErrorContext(contex, "could not marshal user", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(userJSON)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, consts.ApplicationJSONContentType)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (uh *Handler) Self(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	at := ctx.Request.Header.Cookie(consts.ATCookieKey)

	if at == nil {
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	rt := ctx.Request.Header.Cookie(consts.RTCookieKey)

	if rt == nil {
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
	}

	id, err := uh.authUsecase.GetUserID(contex, string(at), string(rt))

	if err != nil {
		uh.logger.WarnContext(contex, "self: getUserID fail", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	user, err := uh.userUsecase.Get(contex, id.UserID)

	if err != nil {
		uh.logger.ErrorContext(contex, "error getting self user", 
			slog.String("ID", id.UserID),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userJSON, err := user.MarshalJSON()

	if err != nil {
		uh.logger.ErrorContext(contex, "could not marshal user", slog.String(consts.ErrorLoggerKey, err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	if id.AccessToken != "" && id.RefreshToken != "" {
		cookies.SetupAccessCookies(ctx, id.ToTokenDTO())
	}

	ctx.Response.SetBody(userJSON)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, consts.ApplicationJSONContentType)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (uh *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("/{id}", uh.Get)
	group.GET("/self", uh.Self)
}
