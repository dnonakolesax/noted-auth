package user

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/model"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type UserUsecase interface {
	Get(uuid string) (model.User, error)
}

type UserHandler struct {
	userUsecase UserUsecase
}

// GerUser godoc
// @Summary Get user info
// @Description Returns user's name, surname and username
// @Tags openid-connect
// @Param id path string true "User ID"
// @Produces json
// @Success 200 {object} model.User
// @Failure 400
// @Failure 500
// @Router /users/{id} [get]
func (uh *UserHandler) Get(ctx *fasthttp.RequestCtx) {
	userId := ctx.UserValue("id")

	if userId == nil {
		slog.Warn("empty user id")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	user, err := uh.userUsecase.Get(userId.(string))

	if err != nil {
		slog.Warn(fmt.Errorf("could not get user: %w", err).Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userJSON, err := json.Marshal(user)

	if err != nil {
		slog.Error(fmt.Errorf("could not marshal user: %w", err).Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(userJSON)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, "application/json")
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (uh *UserHandler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("/{id}", uh.Get)
}

func NewUserHandler(userUsecase UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}
