package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type usecase interface {
	GetAuthLink(retunURL string) (string, error)
	GetToken(state string, token string) (model.TokenDTO, error)
	GetLogoutLink() string
}

type Handler struct {
	basicReturnURL string
	requiredPrefix string
	authUsecase    usecase
}

func NewAuthHandler(basicReturnURL string, requiredPrefix string, authUsecase usecase) *Handler {
	return &Handler{
		basicReturnURL: basicReturnURL,
		requiredPrefix: requiredPrefix,
		authUsecase:    authUsecase,
	}
}

// HandleAuth godoc
// @Summary Handle auth redirect to keycloak
// @Description Generate auth link and redirect to keycloak
// @Tags openid-connect
// @Param return_url query string true "Return url"
// @Success 301
// @Failure 400
// @Failure 500
// @Router /openid-connect/auth [get].
func (ah *Handler) handleAuth(ctx *fasthttp.RequestCtx) {
	returnURL := ctx.QueryArgs().Peek("return_url")
	var returnURLString string
	if returnURL == nil {
		slog.Warn("Return url is empty")
		returnURLString = ah.basicReturnURL
	} else {
		returnURLString = string(returnURL)
	}

	if !strings.HasPrefix(returnURLString, ah.requiredPrefix) {
		slog.Warn(fmt.Sprintf("Return url %v is not allowed", returnURLString))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	redirectLink, err := ah.authUsecase.GetAuthLink(returnURLString)

	if err != nil {
		slog.Error(fmt.Sprintf("Unknown error while getting auth link %v", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Redirect(redirectLink, fasthttp.StatusFound)
}

// HandleToken godoc
// @Summary Handle callback from keycloak
// @Description Receives code and state and returns access token and refresh token
// @Tags openid-connect
// @Param state query string true "State that was sent to keycloak"
// @Param code query string true "Access code from keycloak"
// @Success 301
// @Failure 400
// @Failure 500
// @Router /openid-connect/token [get].
func (ah *Handler) handleToken(ctx *fasthttp.RequestCtx) {
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
		if errors.Is(err, errorvals.ErrObjectNotFoundInRepoError) {
			slog.Warn(fmt.Sprintf("State not found for request-id %s",
				slog.Any("requestId", ctx.Request.Header.Peek("X-Request-Id"))))
			ctx.SetStatusCode(fasthttp.StatusRequestTimeout)
			return
		}
		slog.Error(fmt.Sprintf("Unknown error while getting auth link %v", err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue(tokenDTO.AccessToken)
	atCookie.SetMaxAge(tokenDTO.ExpiresIn)
	atCookie.SetHTTPOnly(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue(tokenDTO.RefreshToken)
	rtCookie.SetMaxAge(tokenDTO.RefreshExp)
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)

	ctx.Redirect(tokenDTO.ReturnURL, fasthttp.StatusFound)
}

func (ah *Handler) HandleLogout(ctx *fasthttp.RequestCtx) {
	ctx.Redirect(ah.authUsecase.GetLogoutLink(), fasthttp.StatusFound)
}

func (ah *Handler) RegisterRoutes(apiGroup *router.Group) {
	group := apiGroup.Group("/openid-connect")
	group.GET("/auth", ah.handleAuth)
	group.GET("/token", ah.handleToken)
}
