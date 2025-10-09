package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	Get(uuid string) (model.User, error)
}

type Handler struct {
	userUsecase usecase
}

func NewUserHandler(userUsecase usecase) *Handler {
	return &Handler{
		userUsecase: userUsecase,
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
	userID := ctx.UserValue("id")

	if userID == nil {
		slog.Warn("empty user id")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	idString, ok := userID.(string)

	if !ok {
		slog.Error(errors.New("could not convert user id to string").Error())
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	user, err := uh.userUsecase.Get(idString)

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

func (uh *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/users")
	group.GET("/{id}", uh.Get)
}
