package auth

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/model"
)

type AuthUsecase interface {
	GetAuthLink(retunUrl string) (string, error)
	GetToken(state string, token string) (model.TokenDTO, error)
}

type AuthHandler struct {
	basicReturnURL string
	requiredPrefix string
	authUsecase AuthUsecase
}

func (ah *AuthHandler) handleAuth(ctx *fasthttp.RequestCtx) {
	returnUrl := ctx.QueryArgs().Peek("return_url")
	var returnUrlString string
	if returnUrl == nil {
		slog.Warn("Return url is empty")
		returnUrlString = ah.basicReturnURL
	} else {
		returnUrlString = string(returnUrl)
	}

	if !strings.HasPrefix(returnUrlString, ah.requiredPrefix) {
		slog.Warn(fmt.Sprintf("Return url %v is not allowed", returnUrlString))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	redirectLink, err := ah.authUsecase.GetAuthLink(returnUrlString)

	if err != nil {
		slog.Error(fmt.Sprintf("Unknown error while getting auth link %v", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return 
	}

	ctx.Redirect(redirectLink, fasthttp.StatusMovedPermanently)
}

func (ah *AuthHandler) handleToken(ctx *fasthttp.RequestCtx) {
	state := ctx.QueryArgs().Peek("state")

	if state == nil {
		slog.Warn("State is empty")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	code := ctx.QueryArgs().Peek("code")

	if code == nil {
		slog.Warn("Code is empty")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	tokenDTO, err := ah.authUsecase.GetToken(string(state), string(code))

	if err != nil {
		slog.Error(fmt.Sprintf("Unknown error while getting auth link %v", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return 
	}

	atCookie := fasthttp.Cookie{}
	atCookie.SetKey("at")
	atCookie.SetValue(tokenDTO.AccessToken)
	atCookie.SetMaxAge(tokenDTO.ExpiresIn)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey("rt")
	rtCookie.SetValue(tokenDTO.RefreshToken)
	rtCookie.SetMaxAge(tokenDTO.RefreshExp)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)

	ctx.Redirect(tokenDTO.ReturnURL, fasthttp.StatusMovedPermanently)
}

func (ah *AuthHandler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/openid-connect")
	group.GET("/auth", ah.handleAuth)
	group.GET("/token", ah.handleToken)
}

func NewAuthHandler(basicReturnURL string, requiredPrefix string, authUsecase AuthUsecase) *AuthHandler {
	return &AuthHandler{
		basicReturnURL: basicReturnURL,
		requiredPrefix: requiredPrefix,
		authUsecase: authUsecase,
	}
}
