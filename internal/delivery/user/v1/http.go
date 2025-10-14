package user

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	Get(ctx context.Context, uuid string) (model.User, error)
}

type Handler struct {
	userUsecase usecase
	logger      *slog.Logger
}

func NewUserHandler(userUsecase usecase, logger *slog.Logger) *Handler {
	return &Handler{
		userUsecase: userUsecase,
		logger:      logger,
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
		uh.logger.WarnContext(contex, "could not get user", "error", err.Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userJSON, err := json.Marshal(user)

	if err != nil {
		uh.logger.ErrorContext(contex, "could not marshal user", "error", err.Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(userJSON)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, "application/json")
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (uh *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("/{id}", uh.Get)
}
