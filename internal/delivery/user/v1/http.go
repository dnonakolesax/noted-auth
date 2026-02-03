package user

import (
	"context"
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	Get(ctx context.Context, uuid string) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.UserID, error)
}

type Handler struct {
	userUsecase usecase
	logger      *slog.Logger
	mw          func(h fasthttp.RequestHandler) fasthttp.RequestHandler
}

func NewUserHandler(userUsecase usecase, logger *slog.Logger,
	mwFunc func(h fasthttp.RequestHandler) fasthttp.RequestHandler) *Handler {
	return &Handler{
		userUsecase: userUsecase,
		logger:      logger,
		mw:          mwFunc,
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

	id := ctx.UserValue(consts.CtxUserIDKey)

	userID, ok := id.(string)

	if !ok {
		uh.logger.WarnContext(contex, "self: fail casting id to string", slog.Any("user_id", userID))
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

	user, err := uh.userUsecase.Get(contex, userID)

	if err != nil {
		uh.logger.ErrorContext(contex, "error getting self user",
			slog.String("ID", userID),
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

	ctx.Response.SetBody(userJSON)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, consts.ApplicationJSONContentType)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (uh *Handler) GetByName(ctx *fasthttp.RequestCtx) {
	trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
	contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)
	userName := ctx.UserValue("name")

	if userName == nil {
		uh.logger.WarnContext(contex, "empty user id")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	unameString, ok := userName.(string)

	if !ok {
		uh.logger.ErrorContext(contex, "could not convert user name to string")
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	user, err := uh.userUsecase.GetByUsername(contex, unameString)

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

func (uh *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("/{id}", uh.mw(uh.Get))
	group.GET("/name/{name}", uh.mw(uh.GetByName))
	group.GET("/self", uh.mw(uh.Self))
}
